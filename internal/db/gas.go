package db

func (repo *Repo) InsertGasEstimate(gasEstimate *GasInfo) error {
	return repo.DB.Create(gasEstimate).Error
}
