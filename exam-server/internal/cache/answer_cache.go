package cache

import (
	"fmt"
	"log"
	"sync"
	"time"

	"exam-server/config"
	"exam-server/internal/database"
	"exam-server/internal/model"
)

type answerEntry struct {
	ExamRecordID uint64
	QuestionID   uint64
	Answer       string
	UpdatedAt    time.Time
}

type AnswerCache struct {
	mu    sync.RWMutex
	data  map[string]*answerEntry
	stopCh chan struct{}
}

var GlobalCache *AnswerCache

func Init() {
	GlobalCache = &AnswerCache{
		data:   make(map[string]*answerEntry),
		stopCh: make(chan struct{}),
	}
	go GlobalCache.periodicFlush()
	log.Println("答案缓存初始化完成")
}

func key(recordID, questionID uint64) string {
	return fmt.Sprintf("exam:answer:%d:%d", recordID, questionID)
}

func (ac *AnswerCache) Set(recordID, questionID uint64, answer string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.data[key(recordID, questionID)] = &answerEntry{
		ExamRecordID: recordID,
		QuestionID:   questionID,
		Answer:       answer,
		UpdatedAt:    time.Now(),
	}
}

func (ac *AnswerCache) Get(recordID, questionID uint64) (string, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	entry, ok := ac.data[key(recordID, questionID)]
	if !ok {
		return "", false
	}
	return entry.Answer, true
}

func (ac *AnswerCache) FlushRecord(recordID uint64) {
	ac.mu.Lock()
	prefix := fmt.Sprintf("exam:answer:%d:", recordID)
	toFlush := make([]*answerEntry, 0)
	for k, v := range ac.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			toFlush = append(toFlush, v)
			delete(ac.data, k)
		}
	}
	ac.mu.Unlock()

	ac.saveToDB(toFlush)
}

func (ac *AnswerCache) periodicFlush() {
	interval := time.Duration(config.AppConfig.Cache.FlushInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ac.flush()
		case <-ac.stopCh:
			ac.flush()
			return
		}
	}
}

func (ac *AnswerCache) flush() {
	ac.mu.Lock()
	entries := make([]*answerEntry, 0, len(ac.data))
	for k, v := range ac.data {
		entries = append(entries, v)
		delete(ac.data, k)
	}
	ac.mu.Unlock()

	ac.saveToDB(entries)
}

func (ac *AnswerCache) saveToDB(entries []*answerEntry) {
	for _, e := range entries {
		detail := model.AnswerDetail{
			ExamRecordID:   e.ExamRecordID,
			QuestionID:     e.QuestionID,
			ProvidedAnswer: e.Answer,
		}
		database.DB.Where("exam_record_id = ? AND question_id = ?", e.ExamRecordID, e.QuestionID).
			Assign(model.AnswerDetail{ProvidedAnswer: e.Answer}).
			FirstOrCreate(&detail)
	}
	if len(entries) > 0 {
		log.Printf("已持久化 %d 条答案至 MySQL", len(entries))
	}
}

func (ac *AnswerCache) Stop() {
	close(ac.stopCh)
}
