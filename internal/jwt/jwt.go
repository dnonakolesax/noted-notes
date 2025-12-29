package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func DecodePayload(jwt string) (map[string]interface{}, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("bad jwt format, has %d parts", len(parts))
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error jwt payload decode: %w", err)
	}
	
	payload := make(map[string]interface{})
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("error jwt payload parsing: %w", err)
	}
	return payload, nil
}
