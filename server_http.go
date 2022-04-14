// Веб-приложение на Go: введение в пакет net/http, пример веб-сервера на Go
// https://golang-blog.blogspot.com/2019/02/go-web-app-net-http-package.html
package main
import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
)

type Page struct {
	Title string
	Body []byte
}


func main()  {
	// Функция main начинается с вызова http.HandleFunc, 
	// который сообщает пакету http обрабатывать все корневые 
	// веб запросы ("/") с помощью handler:
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/view/", viewHandler)
	mux.HandleFunc("/edit/", editHandler)
	// mux.HandleFunc("/save/", saveHandler)
	log.Println("Запуск сервера на http://127.0.0.1:8080")
	// Затем он вызывает http.ListenAndServe, указывая, что он 
	// должен прослушивать порт 8080 на любом интерфейсе (":8080").
	// Эта функция будет блокироваться до завершения программы.
	// ListenAndServe всегда возвращает ошибку, поскольку она возвращается 
	// только тогда, когда случилась неожиданная ошибка. 
	// Чтобы записать эту ошибку в лог, мы заключаем вызов функции в log.Fatal.:
	log.Fatal(http.ListenAndServe(":8080", mux))
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

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	// Опять же, обратите внимание на использование _ для игнорирования error, 
	// при возвращении значения из loadPage. Это сделано здесь для простоты и 
	// вообще считается плохой практикой. 
	p, _ := loadPage(title)
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

// Функция editHandler загружает страницу (или, если он не существует, 
// создает пустую структуру Page), и отображает HTML форму.
func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	fmt.Fprintf(w, "<h1>Editing %s</h1>"+
					"<form action=\"/save/%s\" method=\"POST\">"+
					"<textarea name=\"body\">%s</textarea><br>"+
					"<input type=\"submit\" value=\"Save\">"+
					"</form>",
					p.Title, p.Title, p.Body)
}