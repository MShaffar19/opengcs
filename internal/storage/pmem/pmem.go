// +build linux

package pmem

import (
	"context"
	"fmt"
	"os"

	"github.com/Microsoft/opengcs/internal/oc"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"golang.org/x/sys/unix"
)

// Test dependencies
var (
	osMkdirAll  = os.MkdirAll
	osRemoveAll = os.RemoveAll
	unixMount   = unix.Mount
)

// Mount mounts the pmem device at `/dev/pmem<device>` to `target`.
//
// `target` will be created. On mount failure the created `target` will be
// automatically cleaned up.
//
// Note: For now the platform only supports readonly pmem that is assumed to be
// `ext4`.
func Mount(ctx context.Context, device uint32, target string) (err error) {
	_, span := trace.StartSpan(ctx, "pmem::Mount")
	defer span.End()
	defer func() { oc.SetSpanStatus(span, err) }()

	span.AddAttributes(
		trace.Int64Attribute("device", int64(device)),
		trace.StringAttribute("target", target))

	if err := osMkdirAll(target, 0700); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			osRemoveAll(target)
		}
	}()
	source := fmt.Sprintf("/dev/pmem%d", device)
	flags := uintptr(unix.MS_RDONLY)
	if err := unixMount(source, target, "ext4", flags, "noload"); err != nil {
		return errors.Wrapf(err, "failed to mount pmem device %s onto %s", source, target)
	}
	return nil
}
