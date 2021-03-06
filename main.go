// Веб-приложение на Go: введение в пакет net/http, пример веб-сервера на Go
// https://golang-blog.blogspot.com/2019/02/go-web-app-net-http-package.html
package main
import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	// Пакет html/template помогает гарантировать, что только 
	// безопасный и правильно выглядящий HTML генерируется действиями
	// шаблона. Например, он автоматически экранирует знак «больше»
	// (>), заменяя его с помощью &gt;, чтобы убедиться, что данные
	// пользователя не повреждают HTML форму.
	"html/template"
	"regexp"
	"errors"
)

type Page struct {
	Title string
	Body []byte
}

// Функция template.Must - это удобная оболочка, 
// которая паникует когда передано ненулевое значение error, 
// а в противном случае возвращает *Template без изменений. 
// Здесь уместна паника; если шаблоны не могут быть загружены, 
// единственное разумное, что нужно сделать, это выйти из программы.
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

// Функция regexp.MustCompile проанализирует и скомпилирует регулярное 
// выражение и вернет regexp.Regexp. MustCompile отличается от Compile тем, 
// что он вызывает panic, если компиляция выражения не удается, а Compile 
// возвращает error в качестве второго параметра.
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func main()  {
	// Функция main начинается с вызова http.HandleFunc, 
	// который сообщает пакету http обрабатывать все корневые 
	// веб запросы ("/") с помощью handler:
	// Используется функция http.NewServeMux() для инициализации нового рутера, затем
    // функцию "handler" регистрируется как обработчик для URL-шаблона "/".
	// mux := http.NewServeMux()
	http.HandleFunc("/", handler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Println("Запуск сервера на http://127.0.0.1:8080")
	// Затем он вызывает http.ListenAndServe, указывая, что он 
	// должен прослушивать порт 8080 на любом интерфейсе (":8080").
	// Эта функция будет блокироваться до завершения программы.
	// ListenAndServe всегда возвращает ошибку, поскольку она возвращается 
	// только тогда, когда случилась неожиданная ошибка. 
	// Чтобы записать эту ошибку в лог, мы заключаем вызов функции в log.Fatal.:
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Функция handler имеет тип http.HandlerFunc. 
// Он принимает http.ResponseWriter и http.Request как его аргументы.
// Значение http.ResponseWriter собирает ответ HTTP-сервера; 
// написав в него, мы отправляем данные HTTP-клиенту.
// http.Request - это структура данных, которая представляет клиентский HTTP-запрос.
func handler(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path является компонентом пути URL запроса. 
	// Конечный [1:] означает "создать под-срез Path от 1-го символа до конца."
	// Это удаляет ведущий "/" из имени пути.
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	// w.Write([]byte("Hi there, I love Go!"))
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Опять же, обратите внимание на использование _ для игнорирования error, 
	// при возвращении значения из loadPage. Это сделано здесь для простоты и 
	// вообще считается плохой практикой. 
	p, err := loadPage(title)
	if err != nil {
		// Функция http.Redirect добавляет код статуса HTTP http.StatusFound(302) и
		// Location заголовок к HTTP ответу.
		http.Redirect(w, r, "/edit/"+ title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

// Функция editHandler загружает страницу (или, если он не существует, 
// создает пустую структуру Page), и отображает HTML форму.
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	// Функция template.ParseFiles будет читать содержимое *.html
	// и возвращать *template.Template.
	err :=templates.ExecuteTemplate(w, tmpl + ".html", p)
	if err != nil {
		// Функция http.Error отправляет указанный код HTTP ответа
		// (в данном случае "Internal Server Error") и сообщение об ошибке. 
		// Решение о том, чтобы поместить обработку шаблонов в 
		// отдельную функцию, уже окупается.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Метод t.Execute выполняет шаблон, записывая сгенерированный
	// HTML для http.ResponseWriter. Точечные идентификаторы .Title
	// и .Body относятся к p.Title и p.Body. 
	// err = t.Execute(w, p)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
}

// Функция saveHandler будет обрабатывать отправку форм, 
// которые находятся на страницах редактирования.
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Заголовок страницы (указан в URL) и единственное поле формы, 
	// Body хранятся на новой Page. Затем вызывается метод save() 
	// для записи данных в файл, и клиент перенаправляется на страницу /view/.
	body := r.FormValue("body")
	// Значение, возвращаемое FormValue, имеет тип string. 
	// Мы должны преобразовать это значение в []byte, прежде 
	// чем оно уместится в структуре Page. Мы используем
	// []byte(body) для выполнения преобразования.
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	// О любых ошибках, возникающих во время p.save(), 
	// будет сообщено пользователю.
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

// Теперь давайте напишем функцию, которая использует validPath выражение
// для проверки пути и извлечения заголовка страницы:
// Если заголовок действителен, он будет возвращен вместе с ошибкой со 
// значением nil. Если заголовок недействителен, функция напишет ошибку
// «404 Not Found» для HTTP-соединения и вернет ошибку обработчику. 
func getTitle (w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // title это второе подвыражение.
}

// Теперь давайте определим функцию-обертку, которая принимает 
// функцию верхнего типа и возвращает функцию типа 
// http.HandlerFunc (подходит для передачи в функцию http.HandleFunc):
// Замыкание, возвращаемое makeHandler, является функцией, которая 
// принимает http.ResponseWriter и http.Request (другими словами, 
// http.HandlerFunc). Замыкание извлекает title из пути запроса и 
// проверяет его с помощью TitleValidator regexp. Если title 
// недействителен, ошибка будет записана в ResponseWriter с помощью
// функции http.NotFound. Если title допустим, вложенная 
// функция-обработчик fn будет вызываться с помощью ResponseWriter,
// Request и title в качестве аргументов.
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}