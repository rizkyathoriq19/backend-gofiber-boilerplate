package enum

type ErrorCode int

const (
	// General Success/Info (0-99)
	Success      ErrorCode = 0
	NoDataFound  ErrorCode = 1
	DataNotFound ErrorCode = 2

	// Client Errors (1000-1099)
	InvalidRequest       ErrorCode = -1000
	InvalidRequestBody   ErrorCode = -1001
	MissingRequiredField ErrorCode = -1002
	InvalidFormat        ErrorCode = -1003
	InvalidCredentials   ErrorCode = -1004
	Unauthorized         ErrorCode = -1005
	Forbidden            ErrorCode = -1006
	ResourceNotFound     ErrorCode = -1007
	Conflict             ErrorCode = -1008
	ValidationFailed     ErrorCode = -1009
	InvalidToken         ErrorCode = -1010
	TokenExpired         ErrorCode = -1011
	RateLimitExceeded    ErrorCode = -1012

	// User/Account Errors (1100-1199)
	UsernameExists     ErrorCode = -1100
	EmailExists        ErrorCode = -1101
	InvalidUsername    ErrorCode = -1102
	InvalidEmail       ErrorCode = -1103
	AccountNotFound    ErrorCode = -1104
	AccountInactive    ErrorCode = -1105
	PasswordMismatch   ErrorCode = -1106
	AccountLocked      ErrorCode = -1107
	AccountNotVerified ErrorCode = -1108
	PasswordTooWeak    ErrorCode = -1109

	// File Handling Errors (1200-1299)
	FileSizeExceeded ErrorCode = -1200
	InvalidFileType  ErrorCode = -1201

	// Server Errors (5000-5099)
	InternalServerError  ErrorCode = -5000
	DatabaseError        ErrorCode = -5001
	CacheError           ErrorCode = -5002
	ExternalServiceError ErrorCode = -5003
	ConfigurationError   ErrorCode = -5004
	ServiceUnavailable   ErrorCode = -5005

	// Database Specific (5100-5199)
	DatabaseConnectionFailed ErrorCode = -5100
	DatabaseQueryFailed      ErrorCode = -5101
	DatabaseInsertFailed     ErrorCode = -5102
	DatabaseUpdateFailed     ErrorCode = -5103
	DatabaseDeleteFailed     ErrorCode = -5104
	DatabaseScanFailed       ErrorCode = -5105
	ForeignKeyViolation      ErrorCode = -5106
	TransactionFailed        ErrorCode = -5107

	// Cache/Redis Specific (5200-5299)
	CacheConnectionFailed ErrorCode = -5200
	CacheStoreFailed      ErrorCode = -5201
	CacheRetrieveFailed   ErrorCode = -5202
	CacheDeleteFailed     ErrorCode = -5203

	// Authentication Service (5300-5399)
	TokenGenerationFailed  ErrorCode = -5300
	PasswordHashFailed     ErrorCode = -5301
	AuthServiceUnavailable ErrorCode = -5302

	// File Storage Service (5400-5499)
	FileStorageError ErrorCode = -5400
)

func (ec ErrorCode) Value() int {
	return int(ec)
}

func (ec ErrorCode) String() string {
	names := map[ErrorCode]string{
		// General
		Success:      "SUCCESS",
		NoDataFound:  "NO_DATA_FOUND",
		DataNotFound: "DATA_NOT_FOUND",

		// Client Errors
		InvalidRequest:       "INVALID_REQUEST",
		InvalidRequestBody:   "INVALID_REQUEST_BODY",
		MissingRequiredField: "MISSING_REQUIRED_FIELD",
		InvalidFormat:        "INVALID_FORMAT",
		InvalidCredentials:   "INVALID_CREDENTIALS",
		Unauthorized:         "UNAUTHORIZED",
		Forbidden:            "FORBIDDEN",
		ResourceNotFound:     "RESOURCE_NOT_FOUND",
		Conflict:             "CONFLICT",
		ValidationFailed:     "VALIDATION_FAILED",
		InvalidToken:         "INVALID_TOKEN",
		TokenExpired:         "TOKEN_EXPIRED",
		RateLimitExceeded:    "RATE_LIMIT_EXCEEDED",

		// User/Account Errors
		UsernameExists:     "USERNAME_EXISTS",
		EmailExists:        "EMAIL_EXISTS",
		InvalidUsername:    "INVALID_USERNAME",
		InvalidEmail:       "INVALID_EMAIL",
		AccountNotFound:    "ACCOUNT_NOT_FOUND",
		AccountInactive:    "ACCOUNT_INACTIVE",
		PasswordMismatch:   "PASSWORD_MISMATCH",
		AccountLocked:      "ACCOUNT_LOCKED",
		AccountNotVerified: "ACCOUNT_NOT_VERIFIED",
		PasswordTooWeak:    "PASSWORD_TOO_WEAK",

		// Server Errors
		InternalServerError:  "INTERNAL_SERVER_ERROR",
		DatabaseError:        "DATABASE_ERROR",
		CacheError:           "CACHE_ERROR",
		ExternalServiceError: "EXTERNAL_SERVICE_ERROR",
		ConfigurationError:   "CONFIGURATION_ERROR",
		ServiceUnavailable:   "SERVICE_UNAVAILABLE",

		// Database Specific
		DatabaseConnectionFailed: "DATABASE_CONNECTION_FAILED",
		DatabaseQueryFailed:      "DATABASE_QUERY_FAILED",
		DatabaseInsertFailed:     "DATABASE_INSERT_FAILED",
		DatabaseUpdateFailed:     "DATABASE_UPDATE_FAILED",
		DatabaseDeleteFailed:     "DATABASE_DELETE_FAILED",
		DatabaseScanFailed:       "DATABASE_SCAN_FAILED",
		ForeignKeyViolation:      "FOREIGN_KEY_VIOLATION",
		TransactionFailed:        "TRANSACTION_FAILED",

		// Cache/Redis Specific
		CacheConnectionFailed: "CACHE_CONNECTION_FAILED",
		CacheStoreFailed:      "CACHE_STORE_FAILED",
		CacheRetrieveFailed:   "CACHE_RETRIEVE_FAILED",
		CacheDeleteFailed:     "CACHE_DELETE_FAILED",

		// Authentication Service
		TokenGenerationFailed:  "TOKEN_GENERATION_FAILED",
		PasswordHashFailed:     "PASSWORD_HASH_FAILED",
		AuthServiceUnavailable: "AUTH_SERVICE_UNAVAILABLE",

		// File Handling Errors
		FileSizeExceeded: "FILE_SIZE_EXCEEDED",
		InvalidFileType:  "INVALID_FILE_TYPE",

		// File Storage Service
		FileStorageError: "FILE_STORAGE_ERROR",
	}

	if name, exists := names[ec]; exists {
		return name
	}
	return "UNKNOWN_ERROR"
}

func (ec ErrorCode) MessageID() string {
	messages := map[ErrorCode]string{
		// General
		Success:      "Berhasil",
		NoDataFound:  "Data tidak ditemukan",
		DataNotFound: "Data tidak ditemukan",

		// Client Errors
		InvalidRequest:       "Format permintaan tidak valid",
		InvalidRequestBody:   "Isi permintaan tidak valid",
		MissingRequiredField: "Field wajib tidak lengkap",
		InvalidFormat:        "Format data tidak sesuai",
		InvalidCredentials:   "Kredensial tidak valid",
		Unauthorized:         "Tidak memiliki akses",
		Forbidden:            "Akses ditolak",
		ResourceNotFound:     "Resource tidak ditemukan",
		Conflict:             "Data sudah ada",
		ValidationFailed:     "Validasi gagal",
		InvalidToken:         "Token tidak valid",
		TokenExpired:         "Token sudah kedaluwarsa",
		RateLimitExceeded:    "Batas permintaan terlampaui",

		// User/Account Errors
		UsernameExists:     "Username sudah digunakan",
		EmailExists:        "Email sudah digunakan",
		InvalidUsername:    "Username tidak valid",
		InvalidEmail:       "Format email tidak valid",
		AccountNotFound:    "Akun tidak ditemukan",
		AccountInactive:    "Akun tidak aktif",
		PasswordMismatch:   "Password tidak cocok",
		AccountLocked:      "Akun Anda terkunci.",
		AccountNotVerified: "Akun Anda belum diverifikasi.",
		PasswordTooWeak:    "Password terlalu lemah.",

		// Server Errors
		InternalServerError:  "Terjadi kesalahan pada server",
		DatabaseError:        "Kesalahan database",
		CacheError:           "Kesalahan cache",
		ExternalServiceError: "Layanan eksternal bermasalah",
		ConfigurationError:   "Kesalahan konfigurasi",
		ServiceUnavailable:   "Layanan tidak tersedia",

		// Database Specific
		DatabaseConnectionFailed: "Koneksi database gagal",
		DatabaseQueryFailed:      "Query database gagal",
		DatabaseInsertFailed:     "Gagal menyimpan data",
		DatabaseUpdateFailed:     "Gagal memperbarui data",
		DatabaseDeleteFailed:     "Gagal menghapus data",
		DatabaseScanFailed:       "Gagal membaca data",
		ForeignKeyViolation:      "Gagal menyimpan data karena relasi tidak valid.",
		TransactionFailed:        "Transaksi database gagal.",

		// Cache/Redis Specific
		CacheConnectionFailed: "Koneksi cache gagal",
		CacheStoreFailed:      "Gagal menyimpan ke cache",
		CacheRetrieveFailed:   "Gagal mengambil dari cache",
		CacheDeleteFailed:     "Gagal menghapus dari cache",

		// Authentication Service
		TokenGenerationFailed:  "Gagal membuat token",
		PasswordHashFailed:     "Gagal enkripsi password",
		AuthServiceUnavailable: "Layanan autentikasi tidak tersedia",

		// File Handling Errors
		FileSizeExceeded: "Ukuran file melebihi batas yang diizinkan.",
		InvalidFileType:  "Tipe file tidak valid.",

		// File Storage Service
		FileStorageError: "Gagal menyimpan file.",
	}

	if message, exists := messages[ec]; exists {
		return message
	}
	return "Kesalahan tidak dikenal"
}

func (ec ErrorCode) MessageEN() string {
	messages := map[ErrorCode]string{
		// General
		Success:      "Success",
		NoDataFound:  "No data found",
		DataNotFound: "Data not found",

		// Client Errors
		InvalidRequest:       "Invalid request format",
		InvalidRequestBody:   "Invalid request body",
		MissingRequiredField: "Missing required field",
		InvalidFormat:        "Invalid data format",
		InvalidCredentials:   "Invalid credentials",
		Unauthorized:         "Unauthorized access",
		Forbidden:            "Access forbidden",
		ResourceNotFound:     "Resource not found",
		Conflict:             "Resource already exists",
		ValidationFailed:     "Validation failed",
		InvalidToken:         "Invalid token",
		TokenExpired:         "Token expired",
		RateLimitExceeded:    "Rate limit exceeded",

		// User/Account Errors
		UsernameExists:     "Username already exists",
		EmailExists:        "Email already exists",
		InvalidUsername:    "Invalid username",
		InvalidEmail:       "Invalid email format",
		AccountNotFound:    "Account not found",
		AccountInactive:    "Account is inactive",
		PasswordMismatch:   "Password mismatch",
		AccountLocked:      "Your account is locked.",
		AccountNotVerified: "Your account has not been verified.",
		PasswordTooWeak:    "Password is too weak.",

		// Server Errors
		InternalServerError:  "Internal server error",
		DatabaseError:        "Database error",
		CacheError:           "Cache error",
		ExternalServiceError: "External service error",
		ConfigurationError:   "Configuration error",
		ServiceUnavailable:   "Service unavailable",

		// Database Specific
		DatabaseConnectionFailed: "Database connection failed",
		DatabaseQueryFailed:      "Database query failed",
		DatabaseInsertFailed:     "Failed to insert data",
		DatabaseUpdateFailed:     "Failed to update data",
		DatabaseDeleteFailed:     "Failed to delete data",
		DatabaseScanFailed:       "Failed to scan data",
		ForeignKeyViolation:      "Failed to save data due to an invalid relationship.",
		TransactionFailed:        "Database transaction failed.",

		// Cache/Redis Specific
		CacheConnectionFailed: "Cache connection failed",
		CacheStoreFailed:      "Failed to store in cache",
		CacheRetrieveFailed:   "Failed to retrieve from cache",
		CacheDeleteFailed:     "Failed to delete from cache",

		// Authentication Service
		TokenGenerationFailed:  "Token generation failed",
		PasswordHashFailed:     "Password hashing failed",
		AuthServiceUnavailable: "Authentication service unavailable",

		// File Handling Errors
		FileSizeExceeded: "File size exceeds the allowed limit.",
		InvalidFileType:  "Invalid file type.",

		// File Storage Service
		FileStorageError: "Failed to store file.",
	}

	if message, exists := messages[ec]; exists {
		return message
	}
	return "Unknown error"
}