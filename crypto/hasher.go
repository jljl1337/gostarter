package crypto

/*
Hasher defines the interface for a password hasher. It provides methods for
hashing passwords, comparing hashes, and checking if the hashing parameters
have changed.
*/
type Hasher interface {
	/*
	 Hash generates a hash for the given password string.
	*/
	Hash(password string) (string, error)

	/*
	 Compare checks if the provided hash matches the hash of the password string.
	*/
	Compare(hash, password string) (bool, error)

	/*
	 CompareParameters checks if the parameters of the provided hash match
	 the current hashing parameters. This is used for determining if a hash
	 needs to be rehashed due to changes in hashing parameters.
	*/
	CompareParameters(hash string) (bool, error)
}
