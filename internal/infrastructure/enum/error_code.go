package enum

type ErrorCodeType int

const (
	// General
	NOT_SET_YET ErrorCodeType = iota
	SUCCESS_ON_REQUEST
	NO_DATA_FOUND
	CANT_FETCH_CREATED_DATA

	// Server Errors (range: -10000 to -10999)
	SERVER_ERROR
	SERVER_ERROR_CANT_GENERATE_PASSWORD
	SERVER_ERROR_FAILED_GENERATE_TOKEN
	SERVER_ERROR_REDIS
	SERVER_ERROR_REDIS_CANT_STORE
	SERVER_CANT_INSERT_USER_DATA
	SERVER_CANT_SCAN_USER_DATA

	// Authentication & Authorization (range: -20000 to -20999)
	INVALID_CREDENTIAL
	UNAUTHORIZED
	DONT_HAVE_PERMISSION_TO_ACCESS
	VALIDATED_CREDENTIALS_SUCCESS

	// User & Email Issues (range: -30000 to -30999)
	USERNAME_HAS_BEEN_USED
	EMAIL_HAS_BEEN_USED
	CANT_PICK_USERNAME
	EMAIL_CONTAIN_INVALID_CHARACTER

	// Request & Parsing Errors (range: -40000 to -40999)
	BODY_REQUEST_ERROR
	CANT_PARSE_REQUEST_BODY
	IMPORTANT_BODY_PARSER_NOT_INCLUDED
	UNABLE_GET_PROFILE

	NO_ID_DATA_SEARCH
)

// Value maps each error code to a unique numeric representation
func (enum ErrorCodeType) Value() int {
	return [...]int{
		0,    // NOT_SET_YET
		-100, // SUCCESS_ON_REQUEST
		-101, // NO_DATA_FOUND
		-102, // CANT_FETCH_CREATED_DATA

		-10000, // SERVER_ERROR
		-10010, // SERVER_ERROR_CANT_GENERATE_PASSWORD
		-10020, // SERVER_ERROR_FAILED_GENERATE_TOKEN
		-10030, // SERVER_ERROR_REDIS
		-10031, // SERVER_ERROR_REDIS_CANT_STORE
		-10032, // SERVER_CANT_INSERT_USER_DATA
		-10033, // SERVER_CANT_SCAN_USER_DATA

		-20000, // INVALID_CREDENTIAL
		-20010, // UNAUTHORIZED
		-20020, // DONT_HAVE_PERMISSION_TO_ACCESS
		-20030, // VALIDATED_CREDENTIALS_SUCCESS

		-30000, // USERNAME_HAS_BEEN_USED
		-30010, // EMAIL_HAS_BEEN_USED
		-30020, // CANT_PICK_USERNAME
		-30030, // EMAIL_CONTAIN_INVALID_CHARACTER

		-40000, // BODY_REQUEST_ERROR
		-40010, // CANT_PARSE_REQUEST_BODY
		-40020, // IMPORTANT_BODY_PARSER_NOT_INCLUDED
		-40030, // UNABLE_GET_PROFILE
		-40031, // NO_ID_DATA_SEARCH

	}[enum]
}

// MessageID returns Indonesian messages for each error code
func (enum ErrorCodeType) MessageID() string {
	return [...]string{
		"Kode error belum ditentukan",           // NOT_SET_YET
		"Permintaan berhasil diproses",          // SUCCESS_ON_REQUEST
		"Data tidak ditemukan",                  // NO_DATA_FOUND
		"Gagal mengambil data yang baru dibuat", // CANT_FETCH_CREATED_DATA

		"Terjadi kesalahan pada server, coba lagi nanti", // SERVER_ERROR
		"Gagal membuat kata sandi, coba lagi nanti",      // SERVER_ERROR_CANT_GENERATE_PASSWORD
		"Gagal membuat token, coba lagi nanti",           // SERVER_ERROR_FAILED_GENERATE_TOKEN
		"Kesalahan pada server Redis, coba lagi nanti",   // SERVER_ERROR_REDIS
		"Gagal menyimpan data di Redis",                  // SERVER_ERROR_REDIS_CANT_STORE
		"Gagal menyimpan data pengguna",                  // SERVER_CANT_INSERT_USER_DATA
		"Gagal melakukan analisis data pengguna",         // SERVER_CANT_SCAN_USER_DATA

		"Kredensial tidak valid",              // INVALID_CREDENTIAL
		"Permintaan tidak sah",                // UNAUTHORIZED
		"Tidak memiliki izin untuk mengakses", // DONT_HAVE_PERMISSION_TO_ACCESS
		"Kredensial berhasil divalidasi",      // VALIDATED_CREDENTIALS_SUCCESS

		"Nama pengguna sudah digunakan",                  // USERNAME_HAS_BEEN_USED
		"Email sudah digunakan",                          // EMAIL_HAS_BEEN_USED
		"Tidak dapat menggunakan nama pengguna tersebut", // CANT_PICK_USERNAME
		"Email mengandung karakter tidak valid",          // EMAIL_CONTAIN_INVALID_CHARACTER

		"Format permintaan tidak sesuai", // BODY_REQUEST_ERROR
		"Gagal memproses isi permintaan", // CANT_PARSE_REQUEST_BODY
		"Data penting tidak disertakan",  // IMPORTANT_BODY_PARSER_NOT_INCLUDED
		"Gagal mengambil data profil",    // UNABLE_GET_PROFILE
		"Tidak ada data id pencarian",    // NO_ID_DATA_SEARCH
	}[enum]
}

// MessageEn returns English messages for each error code
func (enum ErrorCodeType) MessageEn() string {
	return [...]string{
		"Error code not set yet",         // NOT_SET_YET
		"Request completed successfully", // SUCCESS_ON_REQUEST
		"No data found",                  // NO_DATA_FOUND
		"Failed to fetch created data",   // CANT_FETCH_CREATED_DATA

		"Server error, please try again later",         // SERVER_ERROR
		"Failed to generate password, try again later", // SERVER_ERROR_CANT_GENERATE_PASSWORD
		"Failed to generate token, try again later",    // SERVER_ERROR_FAILED_GENERATE_TOKEN
		"Redis server error, try again later",          // SERVER_ERROR_REDIS
		"Failed to store data in Redis",                // SERVER_ERROR_REDIS_CANT_STORE
		"Failed store user data",                       // SERVER_CANT_INSERT_USER_DATA
		"Failed to scan user data",                     // SERVER_CANT_SCAN_USER_DATA

		"Invalid credentials provided",         // INVALID_CREDENTIAL
		"Unauthorized request",                 // UNAUTHORIZED
		"Permission denied to access resource", // DONT_HAVE_PERMISSION_TO_ACCESS
		"Credentials successfully validated",   // VALIDATED_CREDENTIALS_SUCCESS

		"Username has already been used",    // USERNAME_HAS_BEEN_USED
		"Email has already been used",       // EMAIL_HAS_BEEN_USED
		"Cannot use the selected username",  // CANT_PICK_USERNAME
		"Email contains invalid characters", // EMAIL_CONTAIN_INVALID_CHARACTER

		"Invalid request format",          // BODY_REQUEST_ERROR
		"Failed to parse request body",    // CANT_PARSE_REQUEST_BODY
		"Important data not included",     // IMPORTANT_BODY_PARSER_NOT_INCLUDED
		"Unable to retrieve profile data", // UNABLE_GET_PROFILE
		"No id search",                    // NO_ID_DATA_SEARCH
	}[enum]
}