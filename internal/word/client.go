package word

import (
	"fmt"

	wordv1 "github.com/hungp29/x-proto/gen/go/word/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewClient dials x-word and returns a WordServiceClient. The returned *grpc.ClientConn
// must be closed by the caller when the application shuts down.
func NewClient(addr string) (wordv1.WordServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("dial x-word at %q: %w", addr, err)
	}
	return wordv1.NewWordServiceClient(conn), conn, nil
}
