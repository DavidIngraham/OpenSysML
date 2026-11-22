//go:build !js || !wasm

package combined

func (s *Server) beginMemoryPhase(string) func() { return nil }

func (s *Server) collectReleasedModel() {}
