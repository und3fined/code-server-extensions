package request

import (
	"log"
	"regexp"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

var fileNameRegex = regexp.MustCompile(`(filename=)([a-zA-Z0-9\.\-]+)`)

type Request struct {
	url     string
	method  string
	body    []byte
	headers map[string]string
	referer string
}

type Response struct {
	Data     []byte
	Filename string
	Size     int
}

func (r *Request) tfsSession() string {
	return uuid.New().String()
}

func (r *Request) request() (Response, error) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.Reset()

	req.SetRequestURI(r.url)
	req.Header.SetMethod(r.method)

	if len(r.body) > 0 {
		req.SetBody(r.body)
	}

	if len(r.headers) > 0 {
		for key, val := range r.headers {
			if key == "Content-Type" {
				req.Header.SetContentType(val)
			} else {
				req.Header.Add(key, val)
			}
		}
	}

	if r.referer != "" {
		req.Header.Add("Referer", r.referer)
	}

	// default header
	req.Header.SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10.16; rv:81.0) Gecko/20100101 Firefox/81.0")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("X-TFS-Session", r.tfsSession())

	resData := Response{
		Filename: "error.txt",
	}

	if err := fasthttp.Do(req, res); err != nil {
		log.Println(err)
		return resData, err
	}

	resData.Size = res.Header.ContentLength()

	if res.Header.StatusCode() == 200 {
		contentDisposition := res.Header.Peek("Content-Disposition")
		if contentDisposition != nil {
			parseFileNameResult := fileNameRegex.FindAllSubmatch(contentDisposition, -1)
			resData.Filename = string(parseFileNameResult[0][2])
		}
	}

	resData.Data = res.Body()

	return resData, nil
}

func (r *Request) SetHeader(key string, value string) *Request {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}

	r.headers[key] = value

	return r
}

func (r *Request) SetBody(body []byte) *Request {
	r.body = body
	return r
}

func (r *Request) ResetBody() *Request {
	r.body = []byte{}
	return r
}

func (r *Request) SetReferer(ref string) *Request {
	r.referer = ref
	return r
}

func (r *Request) Get(url string) (Response, error) {
	r.method = "GET"
	r.url = url

	return r.request()
}

func (r *Request) Post(url string) (Response, error) {
	r.method = "POST"
	r.url = url
	return r.request()
}

func New() *Request {
	return &Request{}
}
