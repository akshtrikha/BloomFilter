package main

import (
	"crypto/rand"
	"fmt"
	"hash"
	"math/big"
	"time"

	"github.com/spaolacci/murmur3"
)

// HashFns - Struct to hold all the hash functions
type HashFns struct {
	mHasherFns []hash.Hash32 // Array of MurmurHash3 Hash Functions
	size       uint32        // Number of Hash Functions
}

// BloomFilter - struct to hold the bloom filter and its size
type BloomFilter struct {
	filter []uint8
	size   uint32
}

// genereate HashFns - function to generate the specified number of hash functions
func generateHashFns(size uint32) *HashFns {
	var mHasherFns []hash.Hash32
	for i := uint32(0); i < size; i++ {
		seed := uint32(time.Now().UnixNano()) + i
		mHasherFns = append(mHasherFns, murmur3.New32WithSeed(seed))
	}

	return &HashFns{mHasherFns, size}
}

// murmurHash - function to generate the hashes and return an array of hash idxs
func murmurHash(key string, filterSize uint32, hashFns *HashFns) []uint32 {
	var idxs []uint32
	sizeHashFns := hashFns.size
	for i := uint32(0); i < sizeHashFns; i++ {
		hashFns.mHasherFns[i].Write([]byte(key))
		idxs = append(idxs, hashFns.mHasherFns[i].Sum32()%filterSize)
		hashFns.mHasherFns[i].Reset()
	}

	// fmt.Println(key, idxs)
	// fmt.Println(idxs)

	return idxs
}

// NewBloomFilter - function that creates the bloom filter
func NewBloomFilter(size uint32) *BloomFilter {
	return &BloomFilter{
		// Taking ceil of the size provided after dividing by 8 to reserve the required memory in bits
		// instead of byts.
		make([]uint8, (size+7)/8),
		size,
	}
}

// Add - method to add key to the bloom filter
func (b *BloomFilter) Add(key string, hashFns *HashFns) {
	/**
	take the key
	pass it into a hash function
	mod the hash with the size
	set that corresponding bit
	*/

	idxs := murmurHash(key, b.size, hashFns)

	for i := 0; i < len(idxs); i++ {
		idx := idxs[i]
		arrayIdx := idx / 8
		bitIdx := idx % 8
		b.filter[arrayIdx] = b.filter[arrayIdx] | (1 << bitIdx)
	}

}

// Print - method to print the bloom filter
func (b *BloomFilter) Print() {
	fmt.Println(b.filter)
}

// Exists = method to check if a key exists in the bloom filter
func (b *BloomFilter) Exists(key string, hashFns *HashFns) bool {
	idxs := murmurHash(key, b.size, hashFns)

	for i := 0; i < len(idxs); i++ {
		idx := idxs[i]

		arrayIdx := idx / 8
		bitIdx := idx % 8
		if !((b.filter[arrayIdx] & (1 << bitIdx)) > 0) {
			return false
		}
	}

	return true
}

func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

func main() {
	var filterSize uint32
	fmt.Println("Enter size of the bloom filter: ")
	fmt.Scanln(&filterSize)
	bloom := NewBloomFilter(filterSize)

	var hashSize uint32
	fmt.Println("Enter the number of hash functions to be used")
	fmt.Scanln(&hashSize)
	hashFns := generateHashFns(hashSize)

	keys := []string{}
	existingKeys := make(map[string]bool)

	// Number of random keys to insert for statistics
	keySize := 50000

	for i := 0; i < keySize; i++ {
		randomKey, _ := generateRandomString(8)
		keys = append(keys, randomKey)
		existingKeys[randomKey] = true
	}

	for _, key := range keys {
		bloom.Add(key, hashFns)
		// bloom.Print()
	}

	// positives := 0
	// for _, key := range keys {
	// 	res := bloom.Exists(key, hashFns)
	// 	fmt.Println(key, res)
	// 	positives++
	// }

	nonExistentKeys := 0
	falsePositives := 0

	for i := 0; i < keySize; i++ {
		randomKey, _ := generateRandomString(8)
		res := bloom.Exists(randomKey, hashFns)

		// fmt.Println(randomKey, res)

		// Counting all the non existing keys for statistics
		if existingKeys[randomKey] == false {
			nonExistentKeys++
		}

		// Counting all the false positives for statistics
		if res && (existingKeys[randomKey] == false) {
			falsePositives++
		}
	}

	fmt.Println("\n\nStatistics:")
	fmt.Println("FilterSize: ", filterSize)
	fmt.Println("HashFnsSize: ", hashSize)
	fmt.Println("FalsePositives: ", falsePositives, "out of", nonExistentKeys, "non existing keys and", keySize, "existing keys.")

}
