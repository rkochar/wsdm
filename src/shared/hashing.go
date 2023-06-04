package shared

import (
	"github.com/google/uuid"
	"hash/crc32"
)

func GetNewID() uuid.UUID {
	return uuid.New()
}

func HashUUID(uuid uuid.UUID) uint32 {
	bytes := uuid[:]
	checksum := crc32.ChecksumIEEE(bytes)

	return checksum % NUM_DBS
}
