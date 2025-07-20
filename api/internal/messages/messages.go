package messages

// Константы для идентификации сервисов
const (
	ServiceEncryption  = "encryption"
	ServiceHealthcheck = "healthcheck"
	ServiceMiddleware  = "middleware"
	ServiceAuth        = "auth"
	ServiceListing     = "listing"
	ServiceStatic      = "static"
)

// Константы для шифрования
const (
	CryptoParamPrime     = "prime"
	CryptoParamGenerator = "generator"
	CryptoSaltedPrefix   = "Salted__"
	CryptoKeyLength      = 32
	CryptoSaltLength     = 8
)

// Ключи для логирования
const (
	LogDetails       = "details"
	LogUserID        = "userID"
	LogSessionID     = "sessionID"
	LogReqPath       = "path"
	LogFilename      = "filename"
	LogKey           = "key"
	LogPrime         = "prime"
	LogGenerator     = "generator"
	LogExpected      = "expected"
	LogGot           = "got"
	LogBlockSize     = "blockSize"
	LogLength        = "length"
	LogUsername      = "username"
	LogListings      = "listings"
	LogTotalPages    = "total_pages"
	LogCurrentPage   = "current_page"
	LogCount         = "count"
	LogTitleLength   = "title_length"
	LogDescLength    = "desc_length"
	LogAddressLength = "address_length"
	LogPrice         = "price"
	LogImageSize     = "image_size"
	LogImageType     = "image_type"
	LogPath          = "path"
	LogID            = "id"
	LogImageURL      = "image_url"
	LogListingID     = "listing_id"
)

// healthcheck
const (
	StatusHealth = "service %s is in state %s"
)

// Поля сортировки
const (
	SortPrice = "price"
	SortDate  = "created_at"
)

// Поля запросов
const (
	ReqUsername     = "username"
	ReqPassword     = "password"
	ReqSortField    = "sort_field"
	ReqSortOrder    = "sort_order"
	ReqOnlyLiked    = "only_liked"
	ReqTargetUserID = "target_user_id"
	ReqPage         = "page"
)

// Токен авторизации
const (
	AuthToken = "AuthToken"
)

// Клиентские ошибки (краткие, понятные пользователю)
const (
	ClientErrAuth                 = "неверное имя пользователя или пароль"
	ClientErrBadRequest           = "некорректный запрос"
	ClientErrNoPermission         = "нет прав доступа"
	ClientErrSessionExpired       = "сессия истекла"
	ClientErrSessionCreation      = "ошибка создания сессии"
	ClientErrUserNotFound         = "пользователь не найден"
	ClientErrInvalidPublicKey     = "некорректный публичный ключ"
	ClientErrDecryption           = "ошибка расшифрования данных"
	ClientErrEncryption           = "ошибка шифрования данных"
	ClientErrPageLoad             = "ошибка загрузки страницы"
	ClientErrNoParams             = "отсутствуют необходимые параметры"
	ClientErrNoCookie             = "требуется авторизация"
	ClientErrBadToken             = "некорректный токен"
	ClientErrNoSession            = "сессия не найдена"
	ClientErrCreateAccount        = "ошибка создания аккаунта"
	ClientErrUserExists           = "пользователь с таким именем уже существует"
	ClientErrNoToken              = "требуется токен авторизации"
	ClientErrInvalidUsername      = "неверный формат логина"
	ClientErrInvalidPass          = "неверный формат пароля"
	ClientErrInvalidUUID          = "неверный формат UUID"
	ClientErrDBQuery              = "ошибка запроса к базе данных"
	ClientErrMissingFields        = "отсутствуют обязательные поля в запросе"
	ClientErrInvalidTitle         = "неверный заголовок объявления"
	ClientErrInvalidDescription   = "неверное описание объявления"
	ClientErrInvalidPrice         = "неверная цена объявления"
	ClientErrInvalidImage         = "неверный формат изображения"
	ClientErrImageTooLarge        = "размер изображения превышает лимит"
	ClientErrUnsupportedImageType = "неподдерживаемый тип изображения"
	ClientErrFileSave             = "ошибка сохранения файла"
	ClientErrInvalidAddress       = "неверный адрес"
	ClientErrMissingID            = "отсутствует ID в запросе"
)

// Логи ошибок (подробные, для отладки)
const (
	LogErrInvalidUsername      = "invalid username format"
	LogErrAuthFailed           = "authentication failed"
	LogErrSessionInvalid       = "invalid session"
	LogErrDBQuery              = "database query failed"
	LogErrEncryption           = "encryption failed"
	LogErrDecryption           = "decryption failed"
	LogErrInvalidPublicKey     = "invalid client public key"
	LogErrKeyDerivation        = "failed to derive shared key"
	LogErrParamsRequest        = "failed to send crypto params"
	LogErrHexDecode            = "failed to decode hex key"
	LogErrKeyLength            = "invalid key length"
	LogErrBase64Decode         = "failed to decode base64 data"
	LogErrMissingSalt          = "missing salt prefix in encrypted data"
	LogErrBlockSize            = "invalid block size"
	LogErrPadding              = "invalid padding"
	LogErrCipherInit           = "failed to initialize cipher"
	LogErrEmptyData            = "received empty data for processing"
	LogErrTokenGeneration      = "failed to generate JWT token"
	LogErrSessionDelete        = "failed to delete session"
	LogErrLoadTemplate         = "failed to load template"
	LogErrRenderTemplate       = "failed to render template"
	LogErrNoAuthToken          = "missing auth token"
	LogErrParseToken           = "failed to parse JWT token"
	LogErrSessionNotFound      = "session not found"
	LogErrUserExists           = "user already exists"
	LogErrInvalidPass          = "invalid password format"
	LogErrInvalidUUID          = "invalid UUID format"
	LogErrMissingFields        = "missing required fields in request"
	LogErrInvalidTitle         = "invalid listing title"
	LogErrInvalidDescription   = "invalid listing description"
	LogErrInvalidPrice         = "invalid listing price"
	LogErrInvalidImage         = "invalid image format"
	LogErrImageTooLarge        = "image size exceeds limit"
	LogErrUnsupportedImageType = "unsupported image type"
	LogErrFileSave             = "failed to save file"
	LogErrInvalidAddress       = "invalid address"
	LogErrMissingID            = "missing ID in request"
)

// Статусы успешных операций для клиента
const (
	StatusSuccess        = "операция выполнена успешно"
	StatusAuth           = "авторизация успешна"
	StatusLogOut         = "выход выполнен"
	StatusListingAdded   = "объявление добавлено успешно"
	StatusListingEdited  = "объявление отредактировано успешно"
	StatusListingDeleted = "объявление удалено успешно"
	StatusLikeAdded      = "лайк добавлен успешно"
	StatusLikeRemoved    = "лайк удалён успешно"
)

// Статусы для логирования успешных операций
const (
	LogStatusUserAuth        = "user authenticated"
	LogStatusUserLogOut      = "user logged out"
	LogStatusParamsSent      = "crypto params sent successfully"
	LogStatusKeyDerived      = "shared key derived successfully"
	LogStatusEncryption      = "data encrypted successfully"
	LogStatusDecryption      = "data decrypted successfully"
	LogStatusPageServed      = "page served successfully"
	LogStatusListingsFetched = "listings fetched successfully"
	LogStatusListingAdded    = "listing added successfully"
	LogStatusListingEdited   = "listing edited successfully"
	LogStatusListingDeleted  = "listing deleted successfully"
	LogStatusLikeAdded       = "like added successfully"
	LogStatusLikeRemoved     = "like removed successfully"
)
