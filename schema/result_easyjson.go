// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package schema

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonD3b49167DecodeGithubComBor3hamRejaSchema(in *jlexer.Lexer, out *Result) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "links":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.Links = make(map[string]*string)
				} else {
					out.Links = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v1 *string
					if in.IsNull() {
						in.Skip()
						v1 = nil
					} else {
						if v1 == nil {
							v1 = new(string)
						}
						*v1 = string(in.String())
					}
					(out.Links)[key] = v1
					in.WantComma()
				}
				in.Delim('}')
			}
		case "data":
			if m, ok := out.Data.(easyjson.Unmarshaler); ok {
				m.UnmarshalEasyJSON(in)
			} else if m, ok := out.Data.(json.Unmarshaler); ok {
				_ = m.UnmarshalJSON(in.Raw())
			} else {
				out.Data = in.Interface()
			}
		case "included":
			if in.IsNull() {
				in.Skip()
				out.Included = nil
			} else {
				if out.Included == nil {
					out.Included = new([]interface{})
				}
				if in.IsNull() {
					in.Skip()
					*out.Included = nil
				} else {
					in.Delim('[')
					if *out.Included == nil {
						if !in.IsDelim(']') {
							*out.Included = make([]interface{}, 0, 4)
						} else {
							*out.Included = []interface{}{}
						}
					} else {
						*out.Included = (*out.Included)[:0]
					}
					for !in.IsDelim(']') {
						var v2 interface{}
						if m, ok := v2.(easyjson.Unmarshaler); ok {
							m.UnmarshalEasyJSON(in)
						} else if m, ok := v2.(json.Unmarshaler); ok {
							_ = m.UnmarshalJSON(in.Raw())
						} else {
							v2 = in.Interface()
						}
						*out.Included = append(*out.Included, v2)
						in.WantComma()
					}
					in.Delim(']')
				}
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD3b49167EncodeGithubComBor3hamRejaSchema(out *jwriter.Writer, in Result) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"links\":")
	if in.Links == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
		out.RawString(`null`)
	} else {
		out.RawByte('{')
		v3First := true
		for v3Name, v3Value := range in.Links {
			if !v3First {
				out.RawByte(',')
			}
			v3First = false
			out.String(string(v3Name))
			out.RawByte(':')
			if v3Value == nil {
				out.RawString("null")
			} else {
				out.String(string(*v3Value))
			}
		}
		out.RawByte('}')
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"data\":")
	if m, ok := in.Data.(easyjson.Marshaler); ok {
		m.MarshalEasyJSON(out)
	} else if m, ok := in.Data.(json.Marshaler); ok {
		out.Raw(m.MarshalJSON())
	} else {
		out.Raw(json.Marshal(in.Data))
	}
	if in.Included != nil {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"included\":")
		if in.Included == nil {
			out.RawString("null")
		} else {
			if *in.Included == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
				out.RawString("null")
			} else {
				out.RawByte('[')
				for v4, v5 := range *in.Included {
					if v4 > 0 {
						out.RawByte(',')
					}
					if m, ok := v5.(easyjson.Marshaler); ok {
						m.MarshalEasyJSON(out)
					} else if m, ok := v5.(json.Marshaler); ok {
						out.Raw(m.MarshalJSON())
					} else {
						out.Raw(json.Marshal(v5))
					}
				}
				out.RawByte(']')
			}
		}
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Result) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD3b49167EncodeGithubComBor3hamRejaSchema(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Result) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD3b49167DecodeGithubComBor3hamRejaSchema(l, v)
}