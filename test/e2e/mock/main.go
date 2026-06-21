// Standalone stdlib-only mock of the Discord REST API, just enough to drive a
// real `terraform apply` + `plan` against the provider for plan-correctness
// validation. Stores objects in memory so reads round-trip with writes.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type store struct {
	mu       sync.Mutex
	nextID   int64
	roles    map[string]map[string]any // guildID -> roleID -> object
	channels map[string]map[string]any // channelID -> object
}

func newStore() *store {
	return &store{nextID: 100000000000000000, roles: map[string]map[string]any{}, channels: map[string]map[string]any{}}
}

func (s *store) id() string {
	s.nextID++
	return strconv.FormatInt(s.nextID, 10)
}

func readJSON(r *http.Request) map[string]any {
	b, _ := io.ReadAll(r.Body)
	m := map[string]any{}
	_ = json.Unmarshal(b, &m)
	return m
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	addr := os.Getenv("MOCK_ADDR")
	if addr == "" {
		addr = "127.0.0.1:17654"
	}
	s := newStore()
	mux := http.NewServeMux()

	// strip the /api/v10 prefix the client adds.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v10")
		parts := strings.Split(strings.Trim(path, "/"), "/")

		s.mu.Lock()
		defer s.mu.Unlock()

		// /guilds/{gid}/roles , /guilds/{gid}/roles/{rid}
		if len(parts) >= 3 && parts[0] == "guilds" && parts[2] == "roles" {
			gid := parts[1]
			if s.roles[gid] == nil {
				s.roles[gid] = map[string]any{}
			}
			switch {
			case r.Method == http.MethodPost && len(parts) == 3:
				body := readJSON(r)
				obj := map[string]any{
					"id":          s.id(),
					"name":        str(body["name"], "new role"),
					"permissions": str(body["permissions"], "0"),
					"color":       num(body["color"]),
					"hoist":       boolean(body["hoist"]),
					"mentionable": boolean(body["mentionable"]),
					"position":    float64(1),
					"managed":     false,
				}
				s.roles[gid][obj["id"].(string)] = obj
				writeJSON(w, 200, obj)
				return
			case r.Method == http.MethodGet && len(parts) == 4:
				if obj, ok := s.roles[gid][parts[3]]; ok {
					writeJSON(w, 200, obj)
					return
				}
				writeJSON(w, 404, map[string]any{"code": 10011, "message": "Unknown Role"})
				return
			case r.Method == http.MethodPatch && len(parts) == 4:
				obj, ok := s.roles[gid][parts[3]]
				if !ok {
					writeJSON(w, 404, map[string]any{"code": 10011, "message": "Unknown Role"})
					return
				}
				body := readJSON(r)
				o := obj.(map[string]any)
				merge(o, body, "name", "permissions", "color", "hoist", "mentionable")
				writeJSON(w, 200, o)
				return
			case r.Method == http.MethodDelete && len(parts) == 4:
				delete(s.roles[gid], parts[3])
				w.WriteHeader(204)
				return
			}
		}

		// /guilds/{gid}/channels (create)
		if len(parts) == 3 && parts[0] == "guilds" && parts[2] == "channels" && r.Method == http.MethodPost {
			body := readJSON(r)
			obj := map[string]any{
				"id":       s.id(),
				"guild_id": parts[1],
				"type":     num(body["type"]),
				"name":     str(body["name"], ""),
				"position": float64(0),
				"nsfw":     boolean(body["nsfw"]),
			}
			if v, ok := body["topic"]; ok {
				obj["topic"] = v
			}
			s.channels[obj["id"].(string)] = obj
			writeJSON(w, 201, obj)
			return
		}

		// /channels/{cid}
		if len(parts) == 2 && parts[0] == "channels" {
			cid := parts[1]
			switch r.Method {
			case http.MethodGet:
				if obj, ok := s.channels[cid]; ok {
					writeJSON(w, 200, obj)
					return
				}
				writeJSON(w, 404, map[string]any{"code": 10003, "message": "Unknown Channel"})
				return
			case http.MethodPatch:
				obj, ok := s.channels[cid]
				if !ok {
					writeJSON(w, 404, map[string]any{"code": 10003, "message": "Unknown Channel"})
					return
				}
				merge(obj, readJSON(r), "name", "topic", "nsfw", "position")
				writeJSON(w, 200, obj)
				return
			case http.MethodDelete:
				delete(s.channels, cid)
				writeJSON(w, 200, map[string]any{"id": cid})
				return
			}
		}

		writeJSON(w, 404, map[string]any{"code": 0, "message": "mock: unhandled " + r.Method + " " + path})
	})

	fmt.Println("mock discord listening on", addr)
	_ = http.ListenAndServe(addr, mux)
}

func str(v any, def string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return def
}
func num(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}
func boolean(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
func merge(dst, src map[string]any, keys ...string) {
	for _, k := range keys {
		if v, ok := src[k]; ok {
			dst[k] = v
		}
	}
}
