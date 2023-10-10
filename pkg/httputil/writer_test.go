package httputil

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestContentTypeDecorator_Decorate(t *testing.T) {
	tests := []struct {
		name string
		d    ContentTypeDecorator
		want string
	}{
		{
			"test1",
			ContentTypeDecorator("testing"),
			"testing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.d.Decorate(w)
			result := w.Result()
			if got := result.Header.Get("content-type"); got != tt.want {
				t.Errorf("ContentTypeDecorator_Decorate() = %v, want %v\n", got, tt.want)
			}
		})
	}
}

func TestCORSDecorator_Decorate(t *testing.T) {
	type fields struct {
		allowedOrigin string
	}
	type want struct {
		allowedOrigin                 string
		accessControlAllowCredentials string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			"test1",
			fields{"www.tokopedia.com"},
			want{"www.tokopedia.com", "true"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			d := &CORSDecorator{
				allowedOrigin: tt.fields.allowedOrigin,
			}
			d.Decorate(w)
			result := w.Result()
			if got := result.Header.Get("access-control-allow-credentials"); got != tt.want.accessControlAllowCredentials {
				t.Errorf("CORSDecorator_Decorate() access-control-allow-credentials = %v, want %v\n", got, tt.want.accessControlAllowCredentials)
				return
			}
			if got := result.Header.Get("access-control-allow-origin"); got != tt.want.allowedOrigin {
				t.Errorf("CORSDecorator_Decorate() access-control-allow-origin = %v, want %v\n", got, tt.want.allowedOrigin)
				return
			}
		})
	}
}

func TestWriteResponse(t *testing.T) {
	type args struct {
		data       []byte
		status     int
		decorators []ResponseDecorator
	}
	type want struct {
		data        []byte
		status      int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"write some bytes",
			args{
				data:   []byte(`this is some bytes`),
				status: http.StatusOK,
			},
			want{
				data:   []byte(`this is some bytes`),
				status: http.StatusOK,
			},
		},
		{
			"with some decorators",
			args{
				data:       []byte(`this is some bytes`),
				status:     http.StatusOK,
				decorators: []ResponseDecorator{NewContentTypeDecorator("application/json")},
			},
			want{
				data:   []byte(`this is some bytes`),
				status: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteResponse(w, tt.args.data, tt.args.status, tt.args.decorators...)
			result := w.Result()
			if data, err := ioutil.ReadAll(result.Body); err != nil {
				t.Errorf("WriteResponse() read from body err = %v\n", err)
				return
			} else {
				if !reflect.DeepEqual(data, tt.want.data) {
					t.Errorf("WriteResponse() body got = %v, want %v\n", data, tt.want.data)
					return
				}
			}
			if tt.want.contentType != "" && result.Header.Get("content-type") != tt.want.contentType {
				t.Errorf("WriteResponse() content-type got = %v, want %v\n", result.Header.Get("content-type"), tt.want.contentType)
				return
			}
		})
	}
}
