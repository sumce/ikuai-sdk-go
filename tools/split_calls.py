#!/usr/bin/env python3
"""把 Burp Suite 导出的 burp.xml 拆分为按 func_name__action 组织的 JSON
文件，供子代理直接阅读（无需解析 XML/base64）。

用法: python -I tools/split_calls.py [burp.xml] [输出目录]
输出: analysis/calls/<func>__<action>.json
      analysis/calls/_index.json  (全部 func/action 清单与文件大小)
"""
import base64
import hashlib
import json
import os
import sys
from collections import OrderedDict
from xml.etree import ElementTree as ET

SRC = sys.argv[1] if len(sys.argv) > 1 else "burp.xml"
OUT = sys.argv[2] if len(sys.argv) > 2 else "analysis/calls"

# 每个 func__action 最多保留的唯一样本数（按 param+响应指纹去重后）。
MAX_SAMPLES_PER_KEY = 10


def b64_or_text(elem):
    if elem is None or elem.text is None:
        return ""
    if (elem.get("base64") or "false").lower() == "true":
        try:
            return base64.b64decode(elem.text).decode("utf-8", errors="replace")
        except Exception:
            return ""
    return elem.text


def split_http(raw):
    if "\r\n\r\n" in raw:
        head, body = raw.split("\r\n\r\n", 1)
    elif "\n\n" in raw:
        head, body = raw.split("\n\n", 1)
    else:
        head, body = raw, ""
    lines = head.split("\r\n") if "\r\n" in head else head.split("\n")
    headers = {}
    status = ""
    if lines and lines[0].startswith("HTTP/"):
        status = lines[0]
        for ln in lines[1:]:
            if ":" in ln:
                k, v = ln.split(":", 1)
                headers[k.strip().lower()] = v.strip()
    return status, headers, body


def main():
    os.makedirs(OUT, exist_ok=True)
    for f in os.listdir(OUT):
        os.remove(os.path.join(OUT, f))
    groups = OrderedDict()  # key -> list of samples
    index = OrderedDict()
    ctx = ET.iterparse(SRC, events=("start", "end"))
    _, root = next(ctx)
    total = 0
    for event, el in ctx:
        if event != "end" or el.tag != "item":
            continue
        total += 1
        url = el.findtext("url") or ""
        req_raw = b64_or_text(el.find("request"))
        resp_raw = b64_or_text(el.find("response"))
        status_line, headers, req_body = split_http(req_raw)
        resp_status, resp_headers, resp_body = split_http(resp_raw)
        # 确定 func/action 键
        if "/Action/call" in url:
            try:
                payload = json.loads(req_body or "{}")
                func = payload.get("func_name", "?")
                action = payload.get("action", "?")
                param = payload.get("param", {})
            except Exception:
                func, action, param = "_unparseable", "?", None
        else:
            func = "_other"
            action = url.rsplit("/", 1)[-1] or "/"
            param = None
        key = f"{func}__{action}"
        # 指纹去重：相同 param + 相同响应只保留一条，避免巨型文件。
        fp = hashlib.sha1(
            (json.dumps(param, sort_keys=True, ensure_ascii=False) +
             "\x00" + resp_body).encode("utf-8", errors="replace")
        ).hexdigest()
        bucket = groups.setdefault(key, OrderedDict())
        if fp not in bucket and len(bucket) < MAX_SAMPLES_PER_KEY:
            bucket[fp] = {
                "url": url,
                "status": resp_status,
                "param": param,
                "request_headers": {
                    k: v for k, v in headers.items()
                    if k in ("cookie", "content-type", "accept", "referer")
                },
                "response": resp_body,
            }
        el.clear()
        if len(root) > 50:
            for child in list(root)[:-1]:
                root.remove(child)

    for key, bucket in groups.items():
        samples = list(bucket.values())
        # Windows 文件名净化：仅保留安全字符。
        safe = "".join(c if c.isalnum() or c in "._-" else "_" for c in key)
        path = os.path.join(OUT, safe + ".json")
        with open(path, "w", encoding="utf-8") as f:
            json.dump(samples, f, ensure_ascii=False, indent=1)
        first = samples[0]
        index[key] = {
            "count": len(samples),
            "file": safe + ".json",
            "bytes": os.path.getsize(path),
            "status": first.get("status", ""),
        }
    with open(os.path.join(OUT, "_index.json"), "w", encoding="utf-8") as f:
        json.dump({"total_items": total, "files": index}, f, ensure_ascii=False, indent=1)
    print(f"parsed {total} items -> {len(groups)} files in {OUT}")
    for k, v in index.items():
        print(f"  {v['count']:3d}x {k}  ({v['bytes']} B)")


if __name__ == "__main__":
    main()
