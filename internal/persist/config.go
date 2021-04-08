package persist

import "time"

var (
	// StorageBucket ...
	StorageBucket = "book"
	// StorageStandard is the standard storage class for general purpose
	StorageStandard StorageClass = "STANDARD"
	storeTimeout                 = time.Second * 50
)

// StorageClass is a Google Cloud storage class
type StorageClass string

func (sc StorageClass) String() string {
	return string(sc)
}
