package session

import (
	"GoMind/common/mysql"
	"GoMind/model"
)

func GetSessionsByUserName(userName string) ([]model.Session, error) {
	var sessions []model.Session
	err := mysql.DB.Where("user_name = ?", userName).Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

func CreateSession(session *model.Session) (*model.Session, error) {
	err := mysql.DB.Create(session).Error
	return session, err
}

func GetSessionByID(sessionID string) (*model.Session, error) {
	var session model.Session
	err := mysql.DB.Where("id = ?", sessionID).First(&session).Error
	return &session, err
}

func DeleteSession(sessionID string) error {
	err := mysql.DB.Where("id = ?", sessionID).Delete(&model.Session{}).Error
	return err
}
