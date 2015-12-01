package player

type PlayerStorage interface {
	ReadPlayer(username string) (*Player, error)

	WritePlayer(player *Player) error
}