package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Service представляет сервис для рекурсивного сохранения веб-страниц с локальными ресурсами.
type Service struct {
	folder     string          // Папка для сохранения файлов
	client     *http.Client    // HTTP клиент для запросов
	parsedUrls map[string]bool // Карта обработанных URL
	depth      uint            // Максимальная глубина рекурсии
}

// NewService создает новый экземпляр Service с указанной папкой и глубиной рекурсии.
func NewService(folder string, depth uint) *Service {
	pagesFolder := filepath.Join("task16", folder)
	err := os.MkdirAll(pagesFolder, 0755)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось создать папку %s: %v\n", pagesFolder, err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Service{
		folder:     pagesFolder,
		client:     client,
		parsedUrls: make(map[string]bool),
		depth:      depth,
	}
}

// fetchHTML получает HTML содержимое по указанному URL.
func (s *Service) fetchHTML(url string) (string, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("ошибка получения URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP ошибка: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	return string(body), nil
}

// Start запускает процесс рекурсивного сохранения веб-страниц начиная с указанного URL.
func (s *Service) Start(url string) error {
	err := s.recursiveParse(url, 0)
	if err != nil {
		return err
	}
	return s.updateLinksInSavedFiles()
}

// recursiveParse рекурсивно парсит веб-страницы до указанной глубины.
func (s *Service) recursiveParse(url string, depth uint) error {
	if depth > s.depth {
		return nil
	}

	if s.parsedUrls[url] {
		return nil
	}

	fmt.Printf("Обрабатываем URL (глубина %d): %s\n", depth, url)

	html, err := s.fetchHTML(url)
	if err != nil {
		fmt.Printf("Ошибка при получении HTML для %s: %v\n", url, err)
		return err
	}

	refactoredHtml, links, err := s.refactorHtml(html, url)
	if err != nil {
		fmt.Printf("Ошибка при рефакторинге HTML для %s: %v\n", url, err)
		return err
	}

	err = s.Save(url, refactoredHtml)
	if err != nil {
		return err
	}

	for _, link := range links {
		if !s.parsedUrls[link] {
			err = s.recursiveParse(link, depth+1)
			if err != nil {
				fmt.Printf("Ошибка при обработке ссылки %s: %v\n", link, err)
				continue
			}
		}
	}

	return nil
}

func (s *Service) refactorHtml(htmlContent string, baseURL string) (res string, links []string, err error) {
	return s.processHTML(htmlContent, baseURL, false)
}

func (s *Service) updateLinksInHTML(htmlContent string, baseURL string) (string, error) {
	result, _, err := s.processHTML(htmlContent, baseURL, true)
	return result, err
}

func (s *Service) processHTML(htmlContent string, baseURL string, updateMode bool) (res string, links []string, err error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", nil, fmt.Errorf("ошибка парсинга HTML: %v", err)
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", nil, fmt.Errorf("ошибка парсинга базового URL: %v", err)
	}

	var extractedLinks []string
	var processNode func(*html.Node)

	processNode = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				for i, attr := range n.Attr {
					if attr.Key == "href" {
						href := s.processLink(attr.Val, base, baseURL, updateMode)
						if href != "" {
							n.Attr[i].Val = href
							if !updateMode && s.isSameDomain(href, baseURL) {
								extractedLinks = append(extractedLinks, href)
							}
						}
					}
				}
			case "img":
				for i, attr := range n.Attr {
					if attr.Key == "src" {
						if src := s.processImage(attr.Val, base, baseURL); src != "" {
							n.Attr[i].Val = src
						}
					}
				}
			case "link":
				for i, attr := range n.Attr {
					if attr.Key == "href" {
						if s.isCSSLink(n) {
							if href := s.processCSS(attr.Val, base, baseURL); href != "" {
								n.Attr[i].Val = href
							}
						}
					}
				}
			case "style":
				s.processInlineStyles(n, base, baseURL)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c)
		}
	}

	processNode(doc)

	var buf strings.Builder
	err = html.Render(&buf, doc)
	if err != nil {
		return "", nil, fmt.Errorf("ошибка рендеринга HTML: %v", err)
	}

	return buf.String(), extractedLinks, nil
}

func (s *Service) Save(url string, html string) error {
	s.parsedUrls[url] = true

	urlFolder := s.generateFolderName(url)
	targetFolder := filepath.Join(s.folder, urlFolder)
	err := os.MkdirAll(targetFolder, 0755)
	if err != nil {
		return fmt.Errorf("ошибка создания папки %s: %v", targetFolder, err)
	}

	fileName := s.generateFileName(url)
	filePath := filepath.Join(targetFolder, fileName)

	err = os.WriteFile(filePath, []byte(html), 0644)
	if err != nil {
		return fmt.Errorf("ошибка сохранения файла %s: %v", filePath, err)
	}

	fmt.Printf("Сохранено: %s\n", filePath)
	return nil
}

// generateFileName создает безопасное имя файла из URL
func (s *Service) generateFileName(urlStr string) string {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return s.sanitizeFileName(urlStr) + ".html"
	}

	fileName := strings.TrimPrefix(parsedUrl.Path, "/")
	if fileName == "" {
		fileName = "index"
	}

	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = s.sanitizeFileName(fileName)

	return fileName + ".html"
}

// sanitizeFileName очищает строку от недопустимых символов для имени файла
func (s *Service) sanitizeFileName(name string) string {
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*", "\\", "/"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}

	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}

	name = strings.Trim(name, "_")

	if name == "" {
		name = "page"
	}

	return name
}

// generateFolderName создает имя папки на основе домена URL
func (s *Service) generateFolderName(urlStr string) string {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return s.sanitizeFileName(urlStr)
	}

	folderName := parsedUrl.Host
	if folderName == "" {
		folderName = "unknown_host"
	}

	folderName = s.sanitizeFileName(folderName)

	return folderName
}

// processLink обрабатывает ссылку универсально
func (s *Service) processLink(href string, base *url.URL, baseURL string, checkLocal bool) string {
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
		if checkLocal {
			return href
		}
		return ""
	}

	if checkLocal && (strings.HasPrefix(href, "./") || strings.HasPrefix(href, "../")) {
		return href
	}

	linkURL, err := url.Parse(href)
	if err != nil {
		return href
	}

	absoluteURL := base.ResolveReference(linkURL)
	absoluteURLStr := absoluteURL.String()

	if checkLocal && s.isSameDomain(absoluteURLStr, baseURL) {
		if localPath := s.getLocalPath(absoluteURLStr, baseURL); localPath != "" {
			return localPath
		}
	}

	return absoluteURLStr
}

// getLocalPath проверяет, существует ли локальный файл для данного URL
func (s *Service) getLocalPath(targetURL string, baseURL string) string {
	if !s.parsedUrls[targetURL] {
		return ""
	}

	targetFolderName := s.generateFolderName(targetURL)
	baseFolderName := s.generateFolderName(baseURL)
	targetFileName := s.generateFileName(targetURL)

	if targetFolderName == baseFolderName {
		return "./" + targetFileName
	}

	return "../" + targetFolderName + "/" + targetFileName
}

// updateLinksInSavedFiles обновляет ссылки во всех сохраненных HTML файлах
func (s *Service) updateLinksInSavedFiles() error {
	fmt.Println("Обновляем ссылки в сохраненных файлах...")

	for pageURL := range s.parsedUrls {
		err := s.updateLinksInFile(pageURL)
		if err != nil {
			fmt.Printf("Предупреждение: не удалось обновить ссылки в файле для %s: %v\n", pageURL, err)
			continue
		}
	}

	fmt.Println("Обновление ссылок завершено!")
	return nil
}

// updateLinksInFile обновляет ссылки в конкретном HTML файле
func (s *Service) updateLinksInFile(pageURL string) error {
	urlFolder := s.generateFolderName(pageURL)
	fileName := s.generateFileName(pageURL)
	filePath := filepath.Join(s.folder, urlFolder, fileName)

	htmlContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла %s: %v", filePath, err)
	}

	updatedHTML, err := s.updateLinksInHTML(string(htmlContent), pageURL)
	if err != nil {
		return fmt.Errorf("ошибка обновления ссылок в HTML: %v", err)
	}

	err = os.WriteFile(filePath, []byte(updatedHTML), 0644)
	if err != nil {
		return fmt.Errorf("ошибка сохранения обновленного файла %s: %v", filePath, err)
	}

	fmt.Printf("Обновлены ссылки в файле: %s\n", filePath)
	return nil
}

// processImage обрабатывает изображение, скачивает его и возвращает локальный путь
func (s *Service) processImage(src string, base *url.URL, baseURL string) string {
	if src == "" {
		return ""
	}

	imgURL, err := url.Parse(src)
	if err != nil {
		return src
	}

	absoluteURL := base.ResolveReference(imgURL)

	localPath := s.downloadImage(absoluteURL.String(), baseURL)
	if localPath != "" {
		return localPath
	}

	return src
}

// downloadResource скачивает ресурс и возвращает локальный путь
func (s *Service) downloadResource(resourceURL string, baseURL string, resourceType string) string {
	urlFolder := s.generateFolderName(baseURL)
	resourceFolder := filepath.Join(s.folder, urlFolder, resourceType)
	err := os.MkdirAll(resourceFolder, 0755)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось создать папку для %s: %v\n", resourceType, err)
		return ""
	}

	parsedURL, err := url.Parse(resourceURL)
	if err != nil {
		return ""
	}

	fileName := filepath.Base(parsedURL.Path)
	if fileName == "" || fileName == "." || fileName == "/" {
		if resourceType == "images" {
			fileName = "image"
		} else {
			fileName = "style"
		}
	}

	fileName = s.sanitizeFileName(fileName)

	if !strings.Contains(fileName, ".") {
		if resourceType == "images" {
			fileName += ".jpg"
		} else {
			fileName += ".css"
		}
	}

	if resourceType == "css" && !strings.HasSuffix(fileName, ".css") {
		fileName += ".css"
	}

	localPath := filepath.Join(resourceFolder, fileName)

	if _, err := os.Stat(localPath); err == nil {
		return "./" + resourceType + "/" + fileName
	}

	resp, err := http.Get(resourceURL)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось скачать %s %s: %v\n", resourceType, resourceURL, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Предупреждение: ошибка при скачивании %s %s: статус %d\n", resourceType, resourceURL, resp.StatusCode)
		return ""
	}

	file, err := os.Create(localPath)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось создать файл %s: %v\n", localPath, err)
		return ""
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось сохранить %s %s: %v\n", resourceType, localPath, err)
		return ""
	}

	fmt.Printf("Скачан %s: %s\n", resourceType, localPath)
	return "./" + resourceType + "/" + fileName
}

// downloadImage скачивает изображение и возвращает локальный путь
func (s *Service) downloadImage(imageURL string, baseURL string) string {
	return s.downloadResource(imageURL, baseURL, "images")
}

// isSameDomain проверяет, принадлежат ли два URL одному домену
func (s *Service) isSameDomain(url1, url2 string) bool {
	u1, err1 := url.Parse(url1)
	u2, err2 := url.Parse(url2)

	if err1 != nil || err2 != nil {
		return false
	}

	return u1.Host == u2.Host
}

// isCSSLink проверяет, является ли тег link ссылкой на CSS файл
func (s *Service) isCSSLink(n *html.Node) bool {
	for _, attr := range n.Attr {
		if attr.Key == "rel" && (attr.Val == "stylesheet" || strings.Contains(attr.Val, "stylesheet")) {
			return true
		}
		if attr.Key == "type" && attr.Val == "text/css" {
			return true
		}
	}
	return false
}

// processCSS обрабатывает CSS файл, скачивает его и возвращает локальный путь
func (s *Service) processCSS(href string, base *url.URL, baseURL string) string {
	if href == "" {
		return ""
	}

	cssURL, err := url.Parse(href)
	if err != nil {
		return href
	}

	absoluteURL := base.ResolveReference(cssURL)

	localPath := s.downloadCSS(absoluteURL.String(), baseURL)
	if localPath != "" {
		return localPath
	}

	return absoluteURL.String()
}

// downloadCSS скачивает CSS файл и возвращает локальный путь
func (s *Service) downloadCSS(cssURL string, baseURL string) string {
	localPath := s.downloadResource(cssURL, baseURL, "css")
	if localPath == "" {
		return ""
	}

	urlFolder := s.generateFolderName(baseURL)
	cssFolder := filepath.Join(s.folder, urlFolder, "css")
	fileName := filepath.Base(cssURL)
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = "style.css"
	}
	fileName = s.sanitizeFileName(fileName)
	if !strings.HasSuffix(fileName, ".css") {
		fileName += ".css"
	}
	localFilePath := filepath.Join(cssFolder, fileName)

	cssContent, err := os.ReadFile(localFilePath)
	if err != nil {
		return localPath
	}

	processedCSS := s.processCSSContent(string(cssContent), cssURL)
	err = os.WriteFile(localFilePath, []byte(processedCSS), 0644)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось обновить CSS %s: %v\n", localFilePath, err)
	}

	return localPath
}

// processCSSContent обрабатывает содержимое CSS файла
func (s *Service) processCSSContent(cssContent string, cssURL string) string {
	base, err := url.Parse(cssURL)
	if err != nil {
		return cssContent
	}

	urlRegex := regexp.MustCompile(`url\s*\(\s*['"]?([^'")]+)['"]?\s*\)`)

	processedCSS := urlRegex.ReplaceAllStringFunc(cssContent, func(match string) string {
		urlMatch := urlRegex.FindStringSubmatch(match)
		if len(urlMatch) < 2 {
			return match
		}

		resourceURL := urlMatch[1]

		if strings.HasPrefix(resourceURL, "data:") || strings.HasPrefix(resourceURL, "http") {
			return match
		}

		parsedURL, err := url.Parse(resourceURL)
		if err != nil {
			return match
		}

		absoluteURL := base.ResolveReference(parsedURL)

		return fmt.Sprintf("url('%s')", absoluteURL.String())
	})

	return processedCSS
}

// processInlineStyles обрабатывает встроенные стили в тегах <style>
func (s *Service) processInlineStyles(n *html.Node, _ *url.URL, baseURL string) {
	if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
		cssContent := n.FirstChild.Data
		processedCSS := s.processCSSContent(cssContent, baseURL)
		n.FirstChild.Data = processedCSS
	}
}
