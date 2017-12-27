package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB
var tpl *template.Template

var fm = template.FuncMap{
	"dollar": toDollar,
}

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://crud:12345@localhost/loja?sslmode=disable")
	if err != nil {
		panic(err)
	}

	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("templates/*.gohtml"))
}

type Produto struct {
	Id        int
	Descricao string
	Dia       int
	Mes       int
	Ano       int
	Imagem    string
	Preco     float32
	Origem    string
	Categoria string
}

type Uniao struct {
	Produtos   []Produto
	Categorias []string
}

func toDollar(real float32) float32 {
	return real / 3.599
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/produtos", produtosIndex)
	http.HandleFunc("/produtos/mostra", mostraProduto)
	http.HandleFunc("/produtos/criar", criaProduto)
	http.HandleFunc("/produtos/criar/processo", criaProdutoProcesso)
	http.HandleFunc("/produtos/update", updateProduto)
	http.HandleFunc("/produtos/update/processo", updateProdutoProcesso)
	http.HandleFunc("/produtos/delete/processo", deleteProdutoProcesso)
	http.ListenAndServe(":50080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/produtos", http.StatusSeeOther)
}

func produtosIndex(w http.ResponseWriter, r *http.Request) {
	categoria := r.FormValue("categoria")
	var query string

	if categoria == "Todos" || categoria == "" {
		query = "SELECT * FROM produtos"
	} else {
		categoria = strings.Join([]string{"'", categoria, "'"}, "")
		query = "SELECT * FROM produtos WHERE categoria"
		query = strings.Join([]string{query, categoria}, "=")
	}

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	pdts := make([]Produto, 0)
	for rows.Next() {
		pdt := Produto{}
		err := rows.Scan(&pdt.Id, &pdt.Descricao, &pdt.Dia, &pdt.Mes, &pdt.Ano, &pdt.Imagem, &pdt.Preco, &pdt.Origem, &pdt.Categoria)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		pdts = append(pdts, pdt)
	}

	rows, err = db.Query("SELECT DISTINCT categoria FROM produtos")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
	}

	cats := make([]string, 0)
	cats = append(cats, "Todos")
	for rows.Next() {
		var cat string
		err := rows.Scan(&cat)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		cats = append(cats, cat)
	}

	uniao := Uniao{
		pdts,
		cats,
	}

	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tpl.ExecuteTemplate(w, "books.gohtml", uniao)
}

func mostraProduto(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")

	row := db.QueryRow("SELECT * FROM produtos WHERE Id = $1", id)

	pdt := Produto{}
	err := row.Scan(&pdt.Id, &pdt.Descricao, &pdt.Dia, &pdt.Mes, &pdt.Ano, &pdt.Imagem, &pdt.Preco, &pdt.Origem, &pdt.Categoria)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(w, "show.gohtml", pdt)
}

func criaProduto(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "create.gohtml", nil)
}

func criaProdutoProcesso(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// pegar valores
	dia, err := strconv.Atoi(r.FormValue("dia"))
	if err != nil {
		panic(err)
	}

	mes, err := strconv.Atoi(r.FormValue("mes"))
	if err != nil {
		panic(err)
	}

	ano, err := strconv.Atoi(r.FormValue("ano"))
	if err != nil {
		panic(err)
	}

	pdt := Produto{}
	pdt.Descricao = r.FormValue("descricao")
	pdt.Dia = dia
	pdt.Mes = mes
	pdt.Ano = ano
	pdt.Origem = r.FormValue("origem")
	pdt.Imagem = r.FormValue("imagem")
	pdt.Categoria = r.FormValue("categoria")
	p := r.FormValue("preco")

	// validar valores
	if pdt.Descricao == "" || pdt.Dia == 0 || pdt.Mes == 0 || pdt.Ano == 0 || pdt.Origem == "" || pdt.Imagem == "" || pdt.Categoria == "" || p == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// converter valores
	f64, err := strconv.ParseFloat(p, 32)
	if err != nil {
		http.Error(w, http.StatusText(406)+"Por favor, retorne e informe um preço válido", http.StatusNotAcceptable)
		return
	}
	pdt.Preco = float32(f64)

	// inserir valores
	_, err = db.Exec("INSERT INTO produtos (descricao, diacompra, mescompra, anocompra, imagem, preco, origem, categoria) VALUES ($1,$2, $3, $4, $5, $6, $7, $8)", pdt.Descricao, pdt.Dia, pdt.Mes, pdt.Ano, pdt.Imagem, pdt.Preco, pdt.Origem, pdt.Categoria)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// confirmar inserção
	tpl.ExecuteTemplate(w, "created.gohtml", pdt)
}

func updateProduto(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM produtos WHERE id = $1", id)

	pdt := Produto{}
	err := row.Scan(&pdt.Id, &pdt.Descricao, &pdt.Dia, &pdt.Mes, &pdt.Ano, &pdt.Imagem, &pdt.Preco, &pdt.Origem, &pdt.Categoria)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "update.gohtml", pdt)
}

func updateProdutoProcesso(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// pegar valores
	dia, err := strconv.Atoi(r.FormValue("dia"))
	if err != nil {
		http.Error(w, http.StatusText(406)+"Por favor, retorne e informe uma data válida", http.StatusNotAcceptable)
		return
	}

	mes, err := strconv.Atoi(r.FormValue("mes"))
	if err != nil {
		http.Error(w, http.StatusText(406)+"Por favor, retorne e informe uma data válida", http.StatusNotAcceptable)
		return
	}

	ano, err := strconv.Atoi(r.FormValue("ano"))
	if err != nil {
		http.Error(w, http.StatusText(406)+"Por favor, retorne e informe uma data válida", http.StatusNotAcceptable)
		return
	}

	pdt := Produto{}
	numero, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, http.StatusText(406)+"Por favor, retorne e informe um id válido", http.StatusNotAcceptable)
		return
	}
	pdt.Id = numero
	pdt.Descricao = r.FormValue("descricao")
	pdt.Dia = dia
	pdt.Mes = mes
	pdt.Ano = ano
	pdt.Origem = r.FormValue("origem")
	pdt.Imagem = r.FormValue("imagem")
	pdt.Categoria = r.FormValue("categoria")
	p := r.FormValue("preco")

	// validar valores
	if pdt.Descricao == "" || pdt.Dia == 0 || pdt.Mes == 0 || pdt.Ano == 0 || pdt.Origem == "" || pdt.Imagem == "" || pdt.Categoria == "" || p == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// converter valores
	f64, err := strconv.ParseFloat(p, 32)
	if err != nil {
		http.Error(w, http.StatusText(406)+"Por favor, retorne e informe um preço válido", http.StatusNotAcceptable)
		return
	}
	pdt.Preco = float32(f64)

	// inserir valores
	_, err = db.Exec("UPDATE produtos SET descricao = $2, diacompra = $3, mescompra = $4, anocompra = $5, imagem = $6, preco = $7, origem = $8, categoria = $9 WHERE id=$1;", pdt.Id, pdt.Descricao, pdt.Dia, pdt.Mes, pdt.Ano, pdt.Imagem, pdt.Preco, pdt.Origem, pdt.Categoria)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// confirmar inserção
	tpl.ExecuteTemplate(w, "updated.gohtml", pdt)
}

func deleteProdutoProcesso(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// deletar livro
	_, err := db.Exec("DELETE FROM produtos WHERE id=$1;", id)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/produtos?categoria=Todos", http.StatusSeeOther)
}
