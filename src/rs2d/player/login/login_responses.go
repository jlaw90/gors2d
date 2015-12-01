package login

type LoginResponseCode uint8

const (
	Begin LoginResponseCode = iota
	RetryInTwoSeconds
	Success
	InvalidUserOrPass
	Banned
	AlreadyLoggedIn
	Updated
	WorldFull
	LoginServerOffline
	TooManyConnectionsFromIP
	BadSessionId
	Rejected
	MembersOnly
	Fail
	Updating
	ReconnectSuccess
	AccountLocked // Too many login attempts
	MembersArea	  // Standing in members area but this is a free server
	InvalidLoginServer LoginResponseCode = 20
	Wait LoginResponseCode = 21 // Just left another world, wait x seconds, where x is a byte sent with the response
)

