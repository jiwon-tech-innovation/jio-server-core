package out

// IntelligencePort Dev 5(Intelligence Worker)와 통신하기 위한 Driven Port
// Emergency 상황에서 AI 분석 요청
type IntelligencePort interface {
	// RequestLogAnalysis 에러 로그 분석 요청 (Emergency Protocol)
	// Dev 6가 EMERGENCY 상태 전송 시 호출
	// 결과는 Markdown 형태로 반환
	RequestLogAnalysis(clientID string, errorLog string, screamText string) (string, error)

	// RequestURLClassification URL/Title을 분석하여 Study vs Play 판별
	RequestURLClassification(clientID string, url string, title string) (string, error)

	// SendAppList: 앱 목록을 전송하고 AI 판정/블랙리스트 결과를 반환
	SendAppList(appsJSON string) (string, string, string, error) // ret: message, command, target_app, error
}
