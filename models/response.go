package models

// EchoResponse represents the complete echo response
type EchoResponse struct {
	Request    RequestInfo        `json:"request"`
	Server     ServerInfo         `json:"server"`
	Kubernetes *KubernetesInfo    `json:"kubernetes,omitempty"`
	JwtTokens  map[string]JwtInfo `json:"jwtTokens,omitempty"`
}

// RequestInfo contains information about the HTTP request
type RequestInfo struct {
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	Query         string            `json:"query,omitempty"`
	Headers       map[string]string `json:"headers"`
	RemoteAddress string            `json:"remoteAddress"`
	Body          *BodyInfo         `json:"body,omitempty"`
	Cookies       []CookieInfo      `json:"cookies,omitempty"`
	Compression   *CompressionInfo  `json:"compression,omitempty"`
	TLS           *RequestTLSInfo   `json:"tls,omitempty"`
}

// BodyInfo contains information about the request body
type BodyInfo struct {
	ContentType string      `json:"contentType,omitempty"`
	Size        int         `json:"size"`
	Content     interface{} `json:"content,omitempty"`
	IsBinary    bool        `json:"isBinary,omitempty"`
	Truncated   bool        `json:"truncated,omitempty"`
}

// ServerInfo contains information about the server
type ServerInfo struct {
	Hostname    string            `json:"hostname"`
	HostAddress string            `json:"hostAddress,omitempty"`
	Environment map[string]string `json:"environment"`
	TLS         *TLSInfo          `json:"tls,omitempty"`
}

// KubernetesInfo contains Kubernetes pod metadata
type KubernetesInfo struct {
	Namespace   string            `json:"namespace"`
	PodName     string            `json:"podName"`
	PodIP       string            `json:"podIp,omitempty"`
	NodeName    string            `json:"nodeName,omitempty"`
	ServiceHost string            `json:"serviceHost,omitempty"`
	ServicePort string            `json:"servicePort,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// JwtInfo contains decoded JWT information
type JwtInfo struct {
	RawToken string                 `json:"rawToken"`
	Header   map[string]interface{} `json:"header,omitempty"`
	Payload  map[string]interface{} `json:"payload,omitempty"`
}

// CookieInfo contains information about an HTTP cookie
type CookieInfo struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Expires  string `json:"expires,omitempty"`
	MaxAge   int    `json:"maxAge,omitempty"`
	HttpOnly bool   `json:"httpOnly,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	SameSite string `json:"sameSite,omitempty"`
}

// CompressionInfo contains information about request/response compression
type CompressionInfo struct {
	AcceptedEncodings []string `json:"acceptedEncodings,omitempty"`
	ResponseEncoding  string   `json:"responseEncoding,omitempty"`
	Supported         bool     `json:"supported"`
}

// TLSInfo contains TLS certificate information for the server
type TLSInfo struct {
	Enabled      bool     `json:"enabled"`
	Version      string   `json:"version,omitempty"`
	Subject      string   `json:"subject,omitempty"`
	Issuer       string   `json:"issuer,omitempty"`
	NotBefore    string   `json:"notBefore,omitempty"`
	NotAfter     string   `json:"notAfter,omitempty"`
	SerialNumber string   `json:"serialNumber,omitempty"`
	DNSNames     []string `json:"dnsNames,omitempty"`
}

// RequestTLSInfo contains TLS information about the specific request
type RequestTLSInfo struct {
	Enabled bool   `json:"enabled"`
	Version string `json:"version,omitempty"`
	Cipher  string `json:"cipher,omitempty"`
}
