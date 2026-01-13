package response

var (
	// Auth messages
	MsgLoginSuccess = BilingualMessage{
		ID: "Login berhasil",
		EN: "Login successful",
	}
	MsgRegisterSuccess = BilingualMessage{
		ID: "User berhasil didaftarkan",
		EN: "User registered successfully",
	}
	MsgLogoutSuccess = BilingualMessage{
		ID: "Logout berhasil",
		EN: "Logout successful",
	}
	MsgTokenRefresh = BilingualMessage{
		ID: "Token berhasil diperbarui",
		EN: "Token refreshed successfully",
	}

	// Profile messages
	MsgProfileRetrieve = BilingualMessage{
		ID: "Profil berhasil diambil",
		EN: "Profile retrieved successfully",
	}
	MsgProfileUpdate = BilingualMessage{
		ID: "Profil berhasil diperbarui",
		EN: "Profile updated successfully",
	}

	// General messages
	MsgSuccess = BilingualMessage{
		ID: "Permintaan berhasil diproses",
		EN: "Request completed successfully",
	}
	MsgDataCreated = BilingualMessage{
		ID: "Data berhasil dibuat",
		EN: "Data created successfully",
	}
	MsgDataRetrieved = BilingualMessage{
		ID: "Data berhasil diambil",
		EN: "Data retrieved successfully",
	}
	MsgDataUpdated = BilingualMessage{
		ID: "Data berhasil diperbarui",
		EN: "Data updated successfully",
	}
	MsgDataDeleted = BilingualMessage{
		ID: "Data berhasil dihapus",
		EN: "Data deleted successfully",
	}
)
