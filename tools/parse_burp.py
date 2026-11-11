#!/usr/bin/env python3
"""Streaming parser for Burp Suite exported XML (burp.xml).

Reads the file in a memory-bounded, chunked manner (xml.etree iterparse +
clear) and extracts every item into a JSONL file plus a summary catalog.
"""
import base64
import json
import sys
from collections import Counter
from urllib.parse import urlparse
from xml.etree import ElementTree as ET

SRC = sys.argv[1] if len(sys.argv) > 1 else "burp.xml"
OUT = sys.argv[2] if len(sys.argv) > 2 else "items.jsonl"
SUMMARY = sys.argv[3] if len(sys.argv) > 3 else "summary.json"


def b64_or_text(elem):
    """Decode a <request>/<response> element: base64 attr or CDATA text."""
    if elem is None or elem.text is None:
        return ""
    if (elem.get("base64") or "false").lower() == "true":
        try:
            return base64.b64decode(elem.text).decode("utf-8", errors="replace")
        except Exception:
            return ""
    return elem.text


def parse_http(raw):
    """Split raw HTTP text into head + body without external libs."""
    if "\r\n\r\n" in raw:
        head, body = raw.split("\r\n\r\n", 1)
    elif "\n\n" in raw:
        head, body = raw.split("\n\n", 1)
    else:
        head, body = raw, ""
    lines = head.split("\r\n") if "\r\n" in head else head.split("\n")
    if lines and lines[0].startswith("HTTP/"):
        # response: status line, then headers
        status = lines[0]
        headers = {}
        for ln in lines[1:]:
            if ":" in ln:
                k, v = ln.split(":", 1)
                headers[k.strip().lower()] = v.strip()
        return {"kind": "response", "status": status, "headers": headers, "body": body}
    if lines:
        parts = lines[0].split(" ")
        method = parts[0] if parts else ""
        target = parts[1] if len(parts) > 1 else ""
        headers = {}
        for ln in lines[1:]:
            if ":" in ln:
                k, v = ln.split(":", 1)
                headers[k.strip().lower()] = v.strip()
        return {"kind": "request", "method": method, "target": target,
                "headers": headers, "body": body}
    return {"kind": "unknown", "headers": {}, "body": raw}


def main():
    items = 0
    url_counter = Counter()
    func_counter = Counter()
    ctx = ET.iterparse(SRC, events=("start", "end"))
    _, root = next(ctx)  # get root element for memory cleanup
    with open(OUT, "w", encoding="utf-8") as out:
        for event, el in ctx:
            if event != "end" or el.tag != "item":
                continue
            url = el.findtext("url") or ""
            host = el.findtext("host") or ""
            port = el.findtext("port") or ""
            protocol = el.findtext("protocol") or ""
            req_raw = b64_or_text(el.find("request"))
            resp_raw = b64_or_text(el.find("response"))
            req = parse_http(req_raw)
            resp = parse_http(resp_raw)
            rec = {
                "url": url, "host": host, "port": port, "protocol": protocol,
                "request": req, "response": resp,
                "status": el.findtext("status") or "",
                "mimetype": el.findtext("mimetype") or "",
                "comment": el.findtext("comment") or "",
            }
            out.write(json.dumps(rec, ensure_ascii=False) + "\n")
            items += 1
            p = urlparse(url)
            url_counter(f"{p.netloc} {p.path}") if False else None
            url_counter[f"{p.netloc}{p.path}"] += 1
            # extract func_name for /Action/call
            if p.path == "/Action/call":
                try:
                    payload = json.loads(req.get("body") or "{}")
                    func_counter[payload.get("func_name", "?") + ":" +
                                 payload.get("action", "?")] += 1
                except Exception:
                    func_counter["<unparseable>"] += 1
            el.clear()
            # free already-processed siblings from the root to bound memory
            if len(root) > 50:
                for child in list(root)[:-1]:
                    root.remove(child)
    summary = {
        "items": items,
        "urls": url_counter.most_common(),
        "func_actions": func_counter.most_common(),
    }
    with open(SUMMARY, "w", encoding="utf-8") as f:
        json.dump(summary, f, ensure_ascii=False, indent=2)
    print(f"parsed {items} items -> {OUT}")
    print(json.dumps(summary, ensure_ascii=False, indent=2)[:4000])


if __name__ == "__main__":
    main()
