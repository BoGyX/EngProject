package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DictionaryService сервис для работы с Free Dictionary API
type DictionaryService struct {
	baseURL string
	client  *http.Client
}

// DictionaryEntry представляет ответ от Dictionary API
type DictionaryEntry struct {
	Word      string    `json:"word"`
	Phonetic  string    `json:"phonetic"`
	Phonetics []Phonetic `json:"phonetics"`
	Meanings  []Meaning  `json:"meanings"`
	Origin    string    `json:"origin"`
}

// Phonetic представляет фонетическую информацию
type Phonetic struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}

// Meaning представляет значение слова
type Meaning struct {
	PartOfSpeech string       `json:"partOfSpeech"`
	Definitions  []Definition `json:"definitions"`
}

// Definition представляет определение слова
type Definition struct {
	Definition string   `json:"definition"`
	Example    string   `json:"example"`
	Synonyms   []string `json:"synonyms"`
	Antonyms   []string `json:"antonyms"`
}

// DictionaryError представляет ошибку от API
type DictionaryError struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Resolution string `json:"resolution"`
}

// NewDictionaryService создает новый сервис для работы с Dictionary API
func NewDictionaryService() *DictionaryService {
	return &DictionaryService{
		baseURL: "https://api.dictionaryapi.dev/api/v2/entries/en",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetWordInfo получает информацию о слове из Dictionary API
func (s *DictionaryService) GetWordInfo(word string) (*DictionaryEntry, error) {
	if word == "" {
		return nil, errors.New("слово не может быть пустым")
	}

	// Формируем URL
	url := fmt.Sprintf("%s/%s", s.baseURL, word)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к Dictionary API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode == http.StatusNotFound {
		var dictErr DictionaryError
		if err := json.Unmarshal(body, &dictErr); err == nil {
			return nil, fmt.Errorf("слово не найдено: %s", dictErr.Message)
		}
		return nil, errors.New("слово не найдено в словаре")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка Dictionary API: статус %d", resp.StatusCode)
	}

	// Парсим ответ (API возвращает массив)
	var entries []DictionaryEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if len(entries) == 0 {
		return nil, errors.New("слово не найдено")
	}

	// Возвращаем первую запись
	return &entries[0], nil
}

// GetWordDefinition получает первое определение слова
func (s *DictionaryService) GetWordDefinition(word string) (string, error) {
	entry, err := s.GetWordInfo(word)
	if err != nil {
		return "", err
	}

	if len(entry.Meanings) == 0 {
		return "", errors.New("определение не найдено")
	}

	if len(entry.Meanings[0].Definitions) == 0 {
		return "", errors.New("определение не найдено")
	}

	return entry.Meanings[0].Definitions[0].Definition, nil
}

// GetWordExample получает первый пример использования слова
func (s *DictionaryService) GetWordExample(word string) (string, error) {
	entry, err := s.GetWordInfo(word)
	if err != nil {
		return "", err
	}

	// Ищем первый пример во всех значениях
	for _, meaning := range entry.Meanings {
		for _, def := range meaning.Definitions {
			if def.Example != "" {
				return def.Example, nil
			}
		}
	}

	return "", errors.New("пример не найден")
}

// GetWordPhonetic получает фонетику слова
func (s *DictionaryService) GetWordPhonetic(word string) (string, error) {
	entry, err := s.GetWordInfo(word)
	if err != nil {
		return "", err
	}

	// Используем основную фонетику или первую из массива
	if entry.Phonetic != "" {
		return entry.Phonetic, nil
	}

	if len(entry.Phonetics) > 0 && entry.Phonetics[0].Text != "" {
		return entry.Phonetics[0].Text, nil
	}

	return "", errors.New("фонетика не найдена")
}

// GetWordAudioURL получает URL аудио произношения
func (s *DictionaryService) GetWordAudioURL(word string) (string, error) {
	entry, err := s.GetWordInfo(word)
	if err != nil {
		return "", err
	}

	// Ищем первое доступное аудио
	for _, phonetic := range entry.Phonetics {
		if phonetic.Audio != "" {
			return phonetic.Audio, nil
		}
	}

	return "", errors.New("аудио не найдено")
}

// MyMemoryTranslationResponse представляет ответ от MyMemory API
type MyMemoryTranslationResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
		Match          float64 `json:"match"`
	} `json:"responseData"`
	ResponseStatus int `json:"responseStatus"`
}

// TranslateToRussian переводит английское слово на русский
func (s *DictionaryService) TranslateToRussian(word string) (string, error) {
	if word == "" {
		return "", errors.New("слово не может быть пустым")
	}

	// Сначала проверяем встроенный словарь для популярных слов
	translations := map[string]string{
		"the": "определенный артикль",
		"power": "сила, мощь, власть",
		"of": "предлог 'из', 'от'",
		"reading": "чтение",
		"is": "есть, является",
		"one": "один",
		"most": "наиболее, самый",
		"important": "важный",
		"skills": "навыки, умения",
		"we": "мы",
		"can": "можем, можно",
		"develop": "развивать",
		"in": "в",
		"our": "наш",
		"lives": "жизни",
		"it": "это, оно",
		"opens": "открывает",
		"doors": "двери",
		"to": "к, в",
		"new": "новый",
		"worlds": "миры",
		"ideas": "идеи",
		"and": "и",
		"perspectives": "перспективы, точки зрения",
		"that": "что, который",
		"might": "может, мог бы",
		"never": "никогда",
		"encounter": "встречать, сталкиваться",
		"otherwise": "иначе, в противном случае",
		"when": "когда",
		"read": "читать",
		"exercise": "упражнять, тренировать",
		"imagination": "воображение",
		"expand": "расширять",
		"vocabulary": "словарный запас",
		"naturally": "естественно",
		"books": "книги",
		"have": "иметь",
		"transport": "переносить, транспортировать",
		"us": "нас",
		"different": "разный, различный",
		"times": "времена",
		"places": "места",
		"through": "через, посредством",
		"experience": "опыт, переживать",
		"life": "жизнь",
		"ancient": "древний",
		"explore": "исследовать",
		"distant": "далекий",
		"galaxies": "галактики",
		"or": "или",
		"understand": "понимать",
		"thoughts": "мысли",
		"people": "люди",
		"from": "из, от",
		"completely": "полностью",
		"cultures": "культуры",
		"this": "это, этот",
		"ability": "способность",
		"see": "видеть",
		"world": "мир",
		"someone": "кто-то",
		"else": "еще, другой",
		"eyes": "глаза",
		"helps": "помогает",
		"empathy": "эмпатия, сопереживание",
		"understanding": "понимание",
		"moreover": "более того, кроме того",
		"improves": "улучшает",
		"cognitive": "когнитивный, познавательный",
		"abilities": "способности",
		"studies": "исследования",
		"shown": "показано",
		"regular": "регулярный",
		"enhance": "улучшать, усиливать",
		"memory": "память",
		"increase": "увеличивать",
		"focus": "фокус, концентрация",
		"even": "даже",
		"reduce": "уменьшать, снижать",
		"stress": "стресс",
		"immerse": "погружать",
		"ourselves": "себя, сами",
		"good": "хороший",
		"book": "книга",
		"brain": "мозг",
		"creates": "создает",
		"neural": "нейронный",
		"pathways": "пути",
		"making": "делая, создавая",
		"more": "более, больше",
		"mentally": "умственно, мысленно",
		"agile": "гибкий, проворный",
		"today": "сегодня",
		"digital": "цифровой",
		"age": "возраст, эпоха",
		"remains": "остается",
		"as": "как, в качестве",
		"relevant": "актуальный, релевантный",
		"ever": "когда-либо",
		"whether": "ли, независимо от того",
		"physical": "физический",
		"articles": "статьи",
		"online": "онлайн, в интернете",
		"act": "акт, действие",
		"continues": "продолжает",
		"be": "быть",
		"fundamental": "фундаментальный, основной",
		"way": "способ, путь",
		"learn": "учиться, изучать",
		"grow": "расти",
		"skill": "навык, умение",
		"serves": "служит",
		"throughout": "на протяжении",
		"entire": "весь, целый",
		"childhood": "детство",
		"education": "образование",
		"professional": "профессиональный",
		"development": "развитие",
		"personal": "личный",
		"enrichment": "обогащение",
		"beauty": "красота",
		"too": "слишком, тоже",
		"late": "поздно",
		"start": "начинать",
		"every": "каждый",
		"adds": "добавляет",
		"knowledge": "знание",
		"so": "так, поэтому",
		"pick": "выбирать, брать",
		"up": "вверх, наверх",
		"discover": "открывать, обнаруживать",
		"endless": "бесконечный",
		"possibilities": "возможности",
		"offer": "предлагать",
	}

	// Проверяем встроенный словарь
	if translation, ok := translations[word]; ok {
		return translation, nil
	}

	// Если слова нет в словаре, используем MyMemory Translation API
	url := fmt.Sprintf("https://api.mymemory.translated.net/get?q=%s&langpair=en|ru", word)
	
	println("Requesting translation for:", word, "URL:", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		println("Error creating request:", err.Error())
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		println("Error making request:", err.Error())
		return "", fmt.Errorf("ошибка запроса к Translation API: %w", err)
	}
	defer resp.Body.Close()

	println("Response status:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка Translation API: статус %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println("Error reading response:", err.Error())
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	println("Response body:", string(body))

	var translationResp MyMemoryTranslationResponse
	if err := json.Unmarshal(body, &translationResp); err != nil {
		println("Error parsing response:", err.Error())
		return "", fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	println("Translation response status:", translationResp.ResponseStatus)
	println("Translated text:", translationResp.ResponseData.TranslatedText)

	if translationResp.ResponseStatus != 200 {
		return "", errors.New("перевод не найден")
	}

	if translationResp.ResponseData.TranslatedText == "" {
		return "", errors.New("перевод не найден")
	}

	return translationResp.ResponseData.TranslatedText, nil
}
