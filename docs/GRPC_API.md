# gRPC API 문서

이 문서는 `jiaa-server-core` 프로젝트의 gRPC API를 정리한 문서입니다.

---

## 목차

1. [SabotageCommandService](#1-sabotagecommandservice)
2. [IntelligenceService (Dev 5)](#2-intelligenceservice-dev-5)
3. [PhysicalControlService (Dev 1)](#3-physicalcontrolservice-dev-1)
4. [ScreenControlService (Dev 3)](#4-screencontrolservice-dev-3)

---

## 1. SabotageCommandService

**Proto 파일:** `api/proto/service.proto`

### ExecuteSabotage

사보타주 명령을 실행합니다.

```protobuf
rpc ExecuteSabotage(SabotageRequest) returns (SabotageResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `action_type` | string | 액션 타입 (BLOCK_URL, CLOSE_APP, MINIMIZE_ALL, SCREEN_GLITCH, RED_FLASH, BLACK_SCREEN, TTS) |
| `intensity` | int32 | 강도 (1-10) |
| `message` | string | 표시할 메시지 |

**Response:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `success` | bool | 성공 여부 |
| `error_code` | string | 에러 코드 |

**사용 예시 (grpcurl):**
```bash
grpcurl -plaintext -d '{
  "client_id": "user-123",
  "action_type": "SCREEN_GLITCH",
  "intensity": 5,
  "message": "집중하세요!"
}' localhost:50051 jiaa.SabotageCommandService/ExecuteSabotage
```

---

## 2. IntelligenceService (Dev 5)

**Proto 파일:** `api/proto/intelligence.proto`

AI 분석 서비스 (STT, URL 분류, 에러 로그 분석)

### AnalyzeLog

에러 로그를 분석하여 해결책을 제시합니다. (Emergency Protocol)

```protobuf
rpc AnalyzeLog(LogAnalysisRequest) returns (LogAnalysisResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `error_log` | string | 에러 로그 텍스트 |
| `scream_text` | string | 비명 텍스트 (STT 결과) |
| `context` | string | 추가 컨텍스트 (선택) |

**Response:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `success` | bool | 성공 여부 |
| `markdown` | string | AI 분석 결과 (Markdown) |
| `solution_code` | string | 해결 코드 (선택) |
| `error_type` | string | 에러 유형 분류 |
| `confidence` | float | 신뢰도 (0.0-1.0) |

---

### ClassifyURL

URL과 제목을 분석하여 분류합니다.

```protobuf
rpc ClassifyURL(URLClassifyRequest) returns (URLClassifyResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `url` | string | URL |
| `title` | string | 페이지 제목 |
| `page_content` | string | 페이지 내용 (선택) |

**Response:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `success` | bool | 성공 여부 |
| `classification` | URLClassification | 분류 결과 |
| `confidence` | float | 신뢰도 (0.0-1.0) |
| `reason` | string | 분류 이유 |

**URLClassification:**
| 값 | 설명 |
|----|------|
| `URL_UNKNOWN` | 알 수 없음 |
| `STUDY` | 공부 관련 |
| `PLAY` | 놀이/오락 관련 |
| `NEUTRAL` | 중립 |
| `WORK` | 업무 관련 |

---

### TranscribeAudio (스트리밍)

실시간 음성을 텍스트로 변환합니다.

```protobuf
rpc TranscribeAudio(stream AudioChunk) returns (TranscribeResponse);
```

**Request (스트리밍):**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `audio_data` | bytes | PCM 오디오 데이터 |
| `sample_rate` | int32 | 샘플 레이트 (예: 16000) |
| `is_final` | bool | 마지막 청크 여부 |

**Response:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `success` | bool | 성공 여부 |
| `text` | string | 변환된 텍스트 |
| `is_final` | bool | 최종 결과 여부 |
| `audio_level` | float | 오디오 레벨 (dB) |

---

## 3. PhysicalControlService (Dev 1)

**Proto 파일:** `api/proto/physical_control.proto`

물리적 제어 서비스 (마우스 감도, 창 흔들기, 앱 종료 등)

### ExecuteCommand

물리적 사보타주 명령을 실행합니다.

```protobuf
rpc ExecuteCommand(PhysicalCommandRequest) returns (PhysicalCommandResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `action_type` | PhysicalActionType | 액션 타입 |
| `intensity` | int32 | 강도 (1-10) |
| `message` | string | 메시지 |
| `params` | map<string, string> | 추가 파라미터 |

**PhysicalActionType:**
| 값 | 설명 |
|----|------|
| `PHYSICAL_ACTION_UNKNOWN` | 알 수 없음 |
| `MOUSE_SENSITIVITY_DOWN` | 마우스 감도 저하 |
| `WINDOW_SHAKE` | 창 흔들기 |
| `KEYBOARD_DELAY` | 키보드 딜레이 |
| `MINIMIZE_ALL` | 모든 창 최소화 |
| `CLOSE_APP` | 앱 종료 |
| `MOUSE_LOCK` | 마우스 잠금 |

**Response:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `success` | bool | 성공 여부 |
| `error_code` | string | 에러 코드 |
| `message` | string | 메시지 |

---

### SetZero

영점 조절 (Calibration Reset)

```protobuf
rpc SetZero(SetZeroRequest) returns (SetZeroResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |

**Response:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `success` | bool | 성공 여부 |

---

## 4. ScreenControlService (Dev 3)

**Proto 파일:** `api/proto/screen_control.proto`

화면 제어 서비스 (시각 효과, AI 결과 표시, TTS 등)

### ExecuteVisualCommand

시각 효과를 실행합니다.

```protobuf
rpc ExecuteVisualCommand(VisualCommandRequest) returns (VisualCommandResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `effect_type` | VisualEffectType | 효과 타입 |
| `intensity` | int32 | 강도 (1-10) |
| `duration_ms` | int32 | 지속 시간 (밀리초) |
| `message` | string | 메시지 |

**VisualEffectType:**
| 값 | 설명 |
|----|------|
| `VISUAL_EFFECT_UNKNOWN` | 알 수 없음 |
| `SCREEN_GLITCH` | 화면 노이즈/글리치 |
| `RED_FLASH` | 붉은 점멸 |
| `SCREEN_SHAKE` | 화면 흔들림 |
| `BLUR_OVERLAY` | 블러 오버레이 |
| `BLACK_SCREEN` | 화면 끔 |
| `VIGNETTE` | 비네팅 효과 |

---

### DisplayAIResult

AI 결과를 Markdown으로 화면에 표시합니다.

```protobuf
rpc DisplayAIResult(AIResultRequest) returns (AIResultResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `markdown` | string | RAG 결과 마크다운 |
| `title` | string | 제목 (선택) |
| `result_type` | AIResultType | 결과 타입 |

**AIResultType:**
| 값 | 설명 |
|----|------|
| `AI_RESULT_UNKNOWN` | 알 수 없음 |
| `ERROR_SOLUTION` | 에러 해결 코드 |
| `TIL_CONTENT` | TIL 콘텐츠 |
| `FACT_BOMB` | 팩트 폭격 |
| `STUDY_TIP` | 공부 팁 |

---

### PlayTTS

TTS를 재생합니다.

```protobuf
rpc PlayTTS(TTSRequest) returns (TTSResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `text` | string | 읽을 텍스트 |
| `voice` | string | 음성 종류 (선택) |
| `speed` | float | 재생 속도 (기본 1.0) |

---

### UpdateScoreGauge

점수 게이지를 업데이트합니다.

```protobuf
rpc UpdateScoreGauge(ScoreUpdateRequest) returns (ScoreUpdateResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `current_score` | int32 | 현재 점수 (0-100) |
| `state` | string | 상태 (THINKING, SLEEPING 등) |

---

### ShowOverlay

오버레이(팝업)를 표시합니다.

```protobuf
rpc ShowOverlay(OverlayRequest) returns (OverlayResponse);
```

**Request:**
| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `title` | string | 제목 |
| `message` | string | 메시지 |
| `overlay_type` | OverlayType | 오버레이 타입 |
| `duration_ms` | int32 | 표시 시간 (0 = 수동 닫기) |

**OverlayType:**
| 값 | 설명 |
|----|------|
| `OVERLAY_UNKNOWN` | 알 수 없음 |
| `WARNING` | 경고 창 |
| `ERROR` | 에러 창 |
| `INFO` | 정보 창 |
| `SUCCESS` | 성공 창 |

---

## 요약

| 서비스 | 대상 | RPC 메서드 수 |
|--------|------|--------------|
| SabotageCommandService | 내부 | 1 |
| IntelligenceService | Dev 5 | 3 |
| PhysicalControlService | Dev 1 | 2 |
| ScreenControlService | Dev 3 | 5 |
| **총합** | - | **11** |
