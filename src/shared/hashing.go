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

func HashTwoUUIDs(uuid1 uuid.UUID, uuid2 uuid.UUID) uint32 {
	hash1 := HashUUID(uuid1)
	hash2 := HashUUID(uuid2)

	return (hash1 + hash2) % NUM_DBS
}
