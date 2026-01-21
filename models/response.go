package models

// EchoResponse represents the complete echo response
type EchoResponse struct {
	Kubernetes *KubernetesInfo    `json:"kubernetes,omitempty"`
	JwtTokens  map[string]JwtInfo `json:"jwtTokens,omitempty"`
	Server     ServerInfo         `json:"server"`
	Request    RequestInfo        `json:"request"`
}

// RequestInfo contains information about the HTTP request
type RequestInfo struct {
	Headers       map[string]string `json:"headers"`
	Body          *BodyInfo         `json:"body,omitempty"`
	Compression   *CompressionInfo  `json:"compression,omitempty"`
	TLS           *RequestTLSInfo   `json:"tls,omitempty"`
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	Query         string            `json:"query,omitempty"`
	RemoteAddress string            `json:"remoteAddress"`
	Cookies       []CookieInfo      `json:"cookies,omitempty"`
}

// BodyInfo contains information about the request body
type BodyInfo struct {
	Content     interface{} `json:"content,omitempty"`
	ContentType string      `json:"contentType,omitempty"`
	Size        int         `json:"size"`
	IsBinary    bool        `json:"isBinary,omitempty"`
	Truncated   bool        `json:"truncated,omitempty"`
}

// ServerInfo contains information about the server
type ServerInfo struct {
	Environment map[string]string `json:"environment"`
	TLS         *TLSInfo          `json:"tls,omitempty"`
	Hostname    string            `json:"hostname"`
	HostAddress string            `json:"hostAddress,omitempty"`
}

// KubernetesInfo contains Kubernetes pod metadata
type KubernetesInfo struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Namespace   string            `json:"namespace"`
	PodName     string            `json:"podName"`
	PodIP       string            `json:"podIp,omitempty"`
	NodeName    string            `json:"nodeName,omitempty"`
	ServiceHost string            `json:"serviceHost,omitempty"`
	ServicePort string            `json:"servicePort,omitempty"`
}

// JwtInfo contains decoded JWT information
type JwtInfo struct {
	Header   map[string]interface{} `json:"header,omitempty"`
	Payload  map[string]interface{} `json:"payload,omitempty"`
	RawToken string                 `json:"rawToken"`
}

// CookieInfo contains information about an HTTP cookie
type CookieInfo struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Expires  string `json:"expires,omitempty"`
	SameSite string `json:"sameSite,omitempty"`
	MaxAge   int    `json:"maxAge,omitempty"`
	HttpOnly bool   `json:"httpOnly,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
}

// CompressionInfo contains information about request/response compression
type CompressionInfo struct {
	ResponseEncoding  string   `json:"responseEncoding,omitempty"`
	AcceptedEncodings []string `json:"acceptedEncodings,omitempty"`
	Supported         bool     `json:"supported"`
}

// TLSInfo contains TLS certificate information for the server
type TLSInfo struct {
	Version      string   `json:"version,omitempty"`
	Subject      string   `json:"subject,omitempty"`
	Issuer       string   `json:"issuer,omitempty"`
	NotBefore    string   `json:"notBefore,omitempty"`
	NotAfter     string   `json:"notAfter,omitempty"`
	SerialNumber string   `json:"serialNumber,omitempty"`
	DNSNames     []string `json:"dnsNames,omitempty"`
	Enabled      bool     `json:"enabled"`
}

// RequestTLSInfo contains TLS information about the specific request
type RequestTLSInfo struct {
	Version string `json:"version,omitempty"`
	Cipher  string `json:"cipher,omitempty"`
	Enabled bool   `json:"enabled"`
}
