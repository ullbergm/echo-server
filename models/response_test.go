package models

import (
	"encoding/json"
	"testing"
)

func TestEchoResponseJSON(t *testing.T) {
	response := EchoResponse{
		Request: RequestInfo{
			Method:        "GET",
			Path:          "/test",
			Query:         "param=value",
			Headers:       map[string]string{"Content-Type": "application/json"},
			RemoteAddress: "127.0.0.1",
		},
		Server: ServerInfo{
			Hostname:    "test-host",
			HostAddress: "192.168.1.1",
			Environment: map[string]string{"ENV": "test"},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal EchoResponse: %v", err)
	}

	// Test JSON unmarshaling
	var decoded EchoResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal EchoResponse: %v", err)
	}

	// Verify fields
	if decoded.Request.Method != response.Request.Method {
		t.Errorf("Expected method %s, got %s", response.Request.Method, decoded.Request.Method)
	}

	if decoded.Server.Hostname != response.Server.Hostname {
		t.Errorf("Expected hostname %s, got %s", response.Server.Hostname, decoded.Server.Hostname)
	}
}

func TestKubernetesInfoJSON(t *testing.T) {
	k8sInfo := KubernetesInfo{
		Namespace:   "default",
		PodName:     "echo-server-123",
		PodIP:       "10.0.0.1",
		NodeName:    "node-1",
		ServiceHost: "echo-server.default.svc.cluster.local",
		ServicePort: "8080",
		Labels: map[string]string{
			"app": "echo-server",
		},
		Annotations: map[string]string{
			"description": "Test pod",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(k8sInfo)
	if err != nil {
		t.Fatalf("Failed to marshal KubernetesInfo: %v", err)
	}

	// Test JSON unmarshaling
	var decoded KubernetesInfo
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal KubernetesInfo: %v", err)
	}

	// Verify fields
	if decoded.Namespace != k8sInfo.Namespace {
		t.Errorf("Expected namespace %s, got %s", k8sInfo.Namespace, decoded.Namespace)
	}

	if decoded.PodName != k8sInfo.PodName {
		t.Errorf("Expected pod name %s, got %s", k8sInfo.PodName, decoded.PodName)
	}

	if len(decoded.Labels) != len(k8sInfo.Labels) {
		t.Errorf("Expected %d labels, got %d", len(k8sInfo.Labels), len(decoded.Labels))
	}
}

func TestJwtInfoJSON(t *testing.T) {
	jwtInfo := JwtInfo{
		RawToken: "eyJhbGc...signature",
		Header: map[string]interface{}{
			"alg": "HS256",
			"typ": "JWT",
		},
		Payload: map[string]interface{}{
			"sub":  "1234567890",
			"name": "John Doe",
			"iat":  1516239022.0,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(jwtInfo)
	if err != nil {
		t.Fatalf("Failed to marshal JwtInfo: %v", err)
	}

	// Test JSON unmarshaling
	var decoded JwtInfo
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal JwtInfo: %v", err)
	}

	// Verify fields
	if decoded.RawToken != jwtInfo.RawToken {
		t.Errorf("Expected raw token %s, got %s", jwtInfo.RawToken, decoded.RawToken)
	}

	if decoded.Header["alg"] != jwtInfo.Header["alg"] {
		t.Errorf("Expected alg %v, got %v", jwtInfo.Header["alg"], decoded.Header["alg"])
	}

	if decoded.Payload["name"] != jwtInfo.Payload["name"] {
		t.Errorf("Expected name %v, got %v", jwtInfo.Payload["name"], decoded.Payload["name"])
	}
}

func TestEchoResponseWithKubernetesInfo(t *testing.T) {
	k8sInfo := &KubernetesInfo{
		Namespace: "production",
		PodName:   "echo-server-abc",
	}

	response := EchoResponse{
		Request: RequestInfo{
			Method: "GET",
			Path:   "/",
		},
		Server: ServerInfo{
			Hostname: "test-host",
		},
		Kubernetes: k8sInfo,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal EchoResponse with K8s info: %v", err)
	}

	var decoded EchoResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal EchoResponse with K8s info: %v", err)
	}

	if decoded.Kubernetes == nil {
		t.Fatal("Expected Kubernetes info to be present")
	}

	if decoded.Kubernetes.Namespace != k8sInfo.Namespace {
		t.Errorf("Expected namespace %s, got %s", k8sInfo.Namespace, decoded.Kubernetes.Namespace)
	}
}

func TestEchoResponseWithJWTTokens(t *testing.T) {
	jwtTokens := map[string]JwtInfo{
		"Authorization": {
			RawToken: "token1",
			Header:   map[string]interface{}{"alg": "HS256"},
			Payload:  map[string]interface{}{"sub": "user1"},
		},
		"X-JWT-Token": {
			RawToken: "token2",
			Header:   map[string]interface{}{"alg": "RS256"},
			Payload:  map[string]interface{}{"sub": "user2"},
		},
	}

	response := EchoResponse{
		Request: RequestInfo{
			Method: "POST",
			Path:   "/api/test",
		},
		Server: ServerInfo{
			Hostname: "api-server",
		},
		JwtTokens: jwtTokens,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal EchoResponse with JWT tokens: %v", err)
	}

	var decoded EchoResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal EchoResponse with JWT tokens: %v", err)
	}

	if len(decoded.JwtTokens) != len(jwtTokens) {
		t.Errorf("Expected %d JWT tokens, got %d", len(jwtTokens), len(decoded.JwtTokens))
	}

	if decoded.JwtTokens["Authorization"].RawToken != "token1" {
		t.Error("Authorization token not decoded correctly")
	}
}

func TestRequestInfoEmptyQuery(t *testing.T) {
	reqInfo := RequestInfo{
		Method:        "GET",
		Path:          "/test",
		Query:         "",
		Headers:       map[string]string{},
		RemoteAddress: "127.0.0.1",
	}

	data, err := json.Marshal(reqInfo)
	if err != nil {
		t.Fatalf("Failed to marshal RequestInfo: %v", err)
	}

	// Verify empty query is handled correctly
	var decoded RequestInfo
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal RequestInfo: %v", err)
	}

	if decoded.Query != "" {
		t.Errorf("Expected empty query, got %s", decoded.Query)
	}
}
