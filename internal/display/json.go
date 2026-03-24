package display

import (
	"encoding/json"
	"fmt"

	"github.com/getkaze/helm/internal/session"
)

func JSON(s *session.Session) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session as JSON: %w", err)
	}
	fmt.Fprintln(Out, string(data))
	return nil
}
