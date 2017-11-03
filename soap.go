package gosoap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"net/http/httputil"
	"log"
)

// Params type is used to set the params in soap request
type Params map[string]interface{}
//type Params map[string]string

// SoapClient return new *Client to handle the requests with the WSDL
func SoapClient(wsdl string) (*Client, error) {
	_, err := url.Parse(wsdl)
	if err != nil {
		return nil, err
	}

	d, err := getWsdlDefinitions(wsdl)
	if err != nil {
		return nil, err
	}

	c := &Client{
		WSDL:        wsdl,
		URL:         strings.TrimSuffix(d.TargetNamespace, "/"),
		Definitions: d,
	}

	return c, nil
}

// Client struct hold all the informations about WSDL,
// request and response of the server
type Client struct {
	debugHTTPRequest    bool
	debugHTTPResponse   bool
	WSDL                string
	URL                 string
	Method              string
	MethodPath          string
	Params              Params
	Definitions         *wsdlDefinitions
	Body                []byte

	payload []byte
}

// Call call's the method m with Params p
func (c *Client) Call(p Params, m ...string) (err error) {
	method, m := m[len(m)-1], m[:len(m)-1]
	path := strings.Join(m, "/")
	if len(path) > 0 {
		path += "/"
	}

	c.Method = method
	c.MethodPath = path
	c.Params = p

	c.payload, err = xml.MarshalIndent(c, "", "")
	if err != nil {
		return err
	}

	b, err := c.doRequest()
	if err != nil {
		return err
	}

	var soap SoapEnvelope
	err = xml.Unmarshal(b, &soap)

	c.Body = soap.Body.Contents

	return err
}

// Unmarshal get the body and unmarshal into the interface
func (c *Client) Unmarshal(v interface{}) error {
	if len(c.Body) == 0 {
		return fmt.Errorf("Body is empty")
	}

	var f Fault
	xml.Unmarshal(c.Body, &f)
	if f.Code != "" {
		return fmt.Errorf("[%s]: %s", f.Code, f.Description)
	}

	return xml.Unmarshal(c.Body, v)
}

// doRequest makes new request to the server using the c.Method, c.URL and the body.
// body is enveloped in Call method
func (c *Client) doRequest() ([]byte, error) {
	req, err := http.NewRequest("POST", c.WSDL, bytes.NewBuffer(c.payload))
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	req.ContentLength = int64(len(c.payload))

	req.Header.Add("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("SOAPAction", fmt.Sprintf("%s/%s%s", c.URL, c.MethodPath, c.Method))
	if (c.debugHTTPRequest) {
		fmt.Println("Dump Request")
		c.debugHTTPOutput(httputil.DumpRequestOut(req, true))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if (c.debugHTTPResponse) {
		fmt.Println("Dump Response")
		c.debugHTTPOutput(httputil.DumpResponse(resp, true))
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// SoapEnvelope struct
type SoapEnvelope struct {
	XMLName struct{} `xml:"Envelope"`
	Body    SoapBody
}

// SoapBody struct
type SoapBody struct {
	XMLName  struct{} `xml:"Body"`
	Contents []byte   `xml:",innerxml"`
}

func (c *Client) EnableResponseDebug() {
	c.debugHTTPResponse = true;
}

func (c *Client) DisableResponseDebug() {
	c.debugHTTPResponse = false;
}

func (c *Client) EnableRequestDebug() {
	c.debugHTTPRequest = true;
}

func (c *Client) DisableRequestDebug() {
	c.debugHTTPRequest = false;
}

func (c *Client) debugHTTPOutput(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}
