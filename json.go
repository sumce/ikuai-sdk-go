package ikuai

import "encoding/json"

// jsonUnmarshal / jsonMarshal 是包内 JSON 辅助别名，便于在多个文件中
// 复用而不必重复 import。
func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }
