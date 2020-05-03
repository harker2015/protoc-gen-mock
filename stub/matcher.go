package stub

import (
	"context"
	"google.golang.org/grpc/metadata"
	"sort"
	"strings"
)

// Search and match stubs in the StubsStore
type StubsMatcher interface {
	Match(ctx context.Context, fullMethod, requestJson string) *Stub
	GetErrorEngine() CustomErrorEngine
}

// Creates new stubs matcher
func NewStubsMatcher(store StubsStore, errorEngine CustomErrorEngine) StubsMatcher {
	return &stubsMatcher{
		StubsStore:        store,
		CustomErrorEngine: errorEngine,
	}
}

type stubsMatcher struct {
	StubsStore        StubsStore
	CustomErrorEngine CustomErrorEngine
}

func (m *stubsMatcher) GetErrorEngine() CustomErrorEngine {
	return m.CustomErrorEngine
}

// Returns the Stub in the StubsStore that matches the method and requestJSON provided OR nil if no stub is found
func (m *stubsMatcher) Match(ctx context.Context, fullMethod, requestJson string) *Stub {
	stubsForMethod := m.StubsStore.GetStubsMapForMethod(fullMethod)
	if stubsForMethod == nil {
		return nil
	}
	for _, stub := range stubsForMethod {
		switch stub.Request.Match {
		case "exact":
			if string(stub.Request.Content) == requestJson && matchMetadata(ctx, stub) {
				return stub
			}
		case "partial":
			// not implemented
		}
	}
	return nil
}

func matchMetadata(ctx context.Context, stub *Stub) bool {
	if len(stub.Request.Metadata) == 0 {
		return true
	}
	// read metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	stubMetadata := getStubMetadata(stub)
	// compare
	for key, values := range stubMetadata {
		contextMetadata := md.Get(key)
		sort.Strings(contextMetadata)
		sort.Strings(values)
		if strings.Join(values, ",") != strings.Join(contextMetadata, ",") {
			return false
		}
	}
	return true
}

func getStubMetadata(stub *Stub) (stubMetadata map[string][]string) {
	stubMetadata = make(map[string][]string, 0)
	for key, value := range stub.Request.Metadata {
		parts := strings.Split(value, ",")
		for _, part := range parts {
			stubMetadata[key] = append(stubMetadata[key], strings.TrimSpace(part))
		}
	}
	return
}
