package gosoap

import (
	"github.com/fatih/structs"
	"encoding/xml"
	"fmt"
	"strconv"
	"reflect"
	"math/big"
)

var tokens []xml.Token

// MarshalXML envelope the body and encode to xml
func (c Client) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	if len(c.Params) == 0 {
		return fmt.Errorf("params is empty")
	}

	tokens = []xml.Token{}

	//start envelope
	if c.Definitions == nil {
		return fmt.Errorf("definitions is nil")
	}

	err := startToken(c.Method, c.Definitions.TargetNamespace)
	if err != nil {
		return err
	}

	for k, v := range c.Params {
		err := deepMarshal(k, v)
		if err != nil {
			return err
		}
	}

	//end envelope
	endToken(c.Method)

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	return e.Flush()
}

func deepMarshal(k string, v interface{}) error {
	ts := xml.StartElement{
		Name: xml.Name{
			Space: "",
			Local: k,
		},
	}

	te := xml.EndElement{Name: ts.Name}

	switch v.(type) {
		case *big.Float:
			bigFloat := *v.(*big.Float)
			tokens = append(tokens, ts, xml.CharData(fmt.Sprintf("%s", bigFloat.String())), te)
			break
		case big.Float:
			bigFloat := v.(big.Float)
			tokens = append(tokens, ts, xml.CharData(fmt.Sprintf("%s", bigFloat.String())), te)
			break
		case *big.Int:
			bigInt := *v.(*big.Int)
			tokens = append(tokens, ts, xml.CharData(fmt.Sprintf("%s", bigInt.String())), te)
			break
		case big.Int:
			bigInt := v.(big.Int)
			tokens = append(tokens, ts, xml.CharData(fmt.Sprintf("%s", bigInt.String())), te)
			break
		case *big.Rat:
			bigRat := *v.(*big.Rat)
			tokens = append(tokens, ts, xml.CharData(fmt.Sprintf("%s", bigRat.FloatString(16))), te)
			break
		case big.Rat:
			bigRat := v.(big.Rat)
			tokens = append(tokens, ts, xml.CharData(fmt.Sprintf("%s", bigRat.FloatString(16))), te)
			break
		case *int:
			if v != nil && v.(*int) != nil {
				tokens = append(tokens, ts, xml.CharData(strconv.Itoa(*v.(*int))), te)
			}
			break
		case int:
			tokens = append(tokens, ts, xml.CharData(strconv.Itoa(v.(int))), te)
			break
		case string:
			if v.(string) == "" {
				break
			}
			tokens = append(tokens, ts, xml.CharData(v.(string)), te)
			break
		case bool:
			tokens = append(tokens, ts, xml.CharData(strconv.FormatBool(v.(bool))), te)
			break
		case map[string]interface{}:
			tokens = append(tokens, ts)
			for dk, dv := range v.(map[string]interface{}) {
				deepMarshal(dk, dv)
			}
			tokens = append(tokens, te)
			break
		case []string:
			tokens = append(tokens, ts)
			for dv := range v.([]string) {
				deepMarshal(k, dv)
			}
			tokens = append(tokens, te)
			break
		case []interface{}:
			tokens = append(tokens, ts)
			for dv := range v.([]interface{}) {
				deepMarshal(k, dv)
			}
			tokens = append(tokens, te)
			break
		case interface{}:
			if (reflect.ValueOf(v).Kind() == reflect.Ptr) {
				break
			}
			tv := structs.Map(v)
			tokens = append(tokens, ts)
			for dk, dv := range tv {
				deepMarshal(dk, dv)
			}
			tokens = append(tokens, te)
			break
		default:
			return fmt.Errorf("UNKNOWN TYPE\nKEY: %s\nTYPE: %s\n", k, reflect.TypeOf(v))
	}

	return nil
}

// startToken initiate body of the envelope
func startToken(m, n string) error {
	e := xml.StartElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Envelope",
		},
		Attr: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "xmlns:xsi"}, Value: "http://www.w3.org/2001/XMLSchema-instance"},
			{Name: xml.Name{Space: "", Local: "xmlns:xsd"}, Value: "http://www.w3.org/2001/XMLSchema"},
			{Name: xml.Name{Space: "", Local: "xmlns:soap"}, Value: "http://schemas.xmlsoap.org/soap/envelope/"},
		},
	}

	b := xml.StartElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Body",
		},
	}

	if m == "" || n == "" {
		return fmt.Errorf("method or namespace is empty")
	}

	r := xml.StartElement{
		Name: xml.Name{
			Space: "",
			Local: m,
		},
		Attr: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "xmlns"}, Value: n},
		},
	}

	tokens = append(tokens, e, b, r)

	return nil
}

// endToken close body of the envelope
func endToken(m string) {
	e := xml.EndElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Envelope",
		},
	}

	b := xml.EndElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Body",
		},
	}

	r := xml.EndElement{
		Name: xml.Name{
			Space: "",
			Local: m,
		},
	}

	tokens = append(tokens, r, b, e)
}
