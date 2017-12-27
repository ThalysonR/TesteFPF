Software feito inteiramente em Golang, sem uso de qualquer framework, desenvolvido em um ambiente unix (Ubuntu 16.04)
Foi utilizado o gerenciador de pacotes go para organizar as dependências.
No front-end, foi utilizado html e css, processados pelo servidor Go.
Foi utilizado o PostgreSQL como banco de dados, segue o script:

CREATE USER crud WITH PASSWORD '12345';
GRANT ALL PRIVILEGES ON DATABASE loja to crud;
ALTER USER crud WITH SUPERUSER;
CREATE DATABASE loja;
\c loja
CREATE TABLE produtos(
ID SERIAL PRIMARY KEY NOT NULL,
DESCRICAO TEXT NOT NULL,
DIACOMPRA INT NOT NULL,
MESCOMPRA INT NOT NULL,
ANOCOMPRA INT NOT NULL,
IMAGEM TEXT NOT NULL,
PRECO REAL NOT NULL,
ORIGEM TEXT NOT NULL,
CATEGORIA TEXT NOT NULL
);

Favor atentar para a ordem de inserção dos campos na tabela, pois a alteração da ordem irá impactar no funcionamento do software.

Código-fonte está contido no arquivo main.go
Foram compilados executáveis para os ambientes unix (Arquivo ServerLinux) e Windows (ServerWindows).
Ao executar o servidor, o serviço irá se iniciar na porta 50080.

Para utilizar o software, deve ser iniciado o banco de dados Postgre e o executável (de acordo com o ambiente) contido neste repositório. Em seguida, deve ser iniciado um navegador, e na barra de endereços, digite: localhost:50080
Favor atentar também para a localização da pasta "templates", que deve estar inserida na pasta que contém o executável do servidor.

