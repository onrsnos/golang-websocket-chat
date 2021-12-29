Chat uygulaması. 

http://localhost:8080/ adresine bağlanacak olan kullanıcılar random isimler alarak birbiri ile mesajlaşabilecekler.

uygulamayı çalıştırabilmek için izlenecek adımlar aşağıdadır.

1-> go mod init main <br>
2-> go get github.com/rsms/gotalk <br>
3-> go run . <br>


uygulamayı build ederek exe çıktısını elde edebilirsiniz

go build

postgresql bağlantısı için <br>
<br>
const (
	host     = "localhost"  
	port     = 5432
	user     = "postgres"
	password = "yourpassword"
	dbname   = "yourdbname"
)<br>

bu kod bloğunu main.go dosyasının 30-37 satırları arasından kendi db nize uygun olarak değiştirebilirsiniz.<br>