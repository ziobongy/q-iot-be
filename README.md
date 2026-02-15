# qiot-configuration-service

Servizio di backend per la gestione delle configurazioni e dei metadati degli esperimenti e dei sensori in un sistema IoT (QiOT).

Contenuti principali
- API REST per: sensori, esperimenti e dashboard.
- Persistenza su MongoDB per configurazioni/metadati.
- Query su InfluxDB per dati temporali dei sensori.
- Semplice configurazione tramite variabili d'ambiente.

Tecnologie
- Linguaggio: Go (>= 1.25)
- Web framework: Gin
- Database: MongoDB (driver ufficiale)
- Time-series DB: InfluxDB (client ufficiale v2)

Requisiti
- Go 1.25+
- MongoDB accessibile (URL via variabile d'ambiente)
- InfluxDB v2 accessibile (URL e token via variabili d'ambiente)

Variabili d'ambiente
- MONGO_URI: URI di connessione a MongoDB (es. mongodb://user:pass@host:27017)
- INFLUX_URI: URL del server InfluxDB (es. https://eu-west-1-1.aws.cloud2.influxdata.com)
- INFLUX_TOKEN: token di autenticazione InfluxDB

Installazione e esecuzione locale
1. Scarica le dipendenze:

   go mod download

2. Imposta le variabili d'ambiente richieste (es. in un terminale zsh):

   export MONGO_URI="mongodb://localhost:27017"
   export INFLUX_URI="https://your-influx-url"
   export INFLUX_TOKEN="your-token"

3. Avvia il servizio:

   go run main.go

Il server gira di default sulla porta 8080 (http://localhost:8080).

Docker
- È presente un `Dockerfile` nella radice del progetto; puoi costruire un'immagine Docker e avviarla in modo tradizionale. Assicurati di fornire le variabili d'ambiente `MONGO_URI`, `INFLUX_URI` e `INFLUX_TOKEN` al container.

Configurazione CORS
- L'applicazione abilita CORS per alcune origini in `main.go` (es. http://localhost:5173 e un dominio remoto) e consente i metodi GET/POST/PUT/PATCH/DELETE/OPTIONS.

Endpoint principali (riassunto)
- GET /sensor
  - Restituisce tutti i sensori registrati.
- GET /sensor/:sensorId
  - Restituisce i dettagli di un sensore (id = sensorId).
- POST /sensor
  - Inserisce una nuova configurazione sensore (body JSON).
- PUT /sensor/:sensorId
  - Aggiorna la configurazione del sensore specificato.
- GET /sensor/:sensorId/characteristic/:serviceUuid
  - Restituisce le caratteristiche associate a un servizio/UUID per il sensore.

- GET /experiment
  - Elenca tutti gli esperimenti.
- GET /experiment/:id
  - Restituisce l'esperimento raw (detail).
- GET /experiment/json/:id
  - Restituisce l'esperimento in JSON (struttura completa).
- GET /experiment/yaml/:id
  - Restituisce l'esperimento in YAML.
- POST /experiment
  - Inserisce un nuovo esperimento (body JSON).
- PUT /experiment/:experimentId
  - Aggiorna un esperimento esistente.

- GET /dashboard/:experimentId/device/:sensorId
  - Endpoint per ottenere dati di dashboard (dati temporali da Influx) per uno specifico sensore/esperimento.

Note sull'architettura
- `config/` contiene i client e la logica di connessione per MongoDB e InfluxDB.
- `api/` espone gli handler HTTP tramite Gin.
- `service/` contiene la logica applicativa che interagisce con i client di `config/`.

Esempi rapidi con curl
- Ottenere tutti i sensori:

   curl -s http://localhost:8080/sensor | jq .

- Inserire un sensore (esempio minimal):

   curl -X POST http://localhost:8080/sensor \
     -H "Content-Type: application/json" \
     -d '{"name":"sensore-1","deviceAddress":"AA:BB:CC:DD","characteristics":[]}'

- Ottenere dati dashboard per sensore:

   curl -s http://localhost:8080/dashboard/<experimentId>/device/<sensorId> | jq .

Buone pratiche e suggerimenti
- Assicurarsi che MongoDB abbia l'indice/collezione `configurations` e, se usati, `experiments`.
- L'ID usato per `experimentId` nelle query verso InfluxDB e negli endpoint deve corrispondere ai metadati salvati insieme alle configurazioni.
- Proteggere l'accesso al token InfluxDB in ambienti di produzione (segreteria, Vault, ecc.).

Contribuire
- Fork & pull request. Mantieni gli standard del progetto e aggiungi test per cambiamenti significativi.

Licenza
- Non è specificata nel repository; aggiungi un file LICENSE se intendi rilasciare il codice con una licenza specifica.

Contatti
- Mantieni il README aggiornato con informazioni di deployment e contatti del team se necessario.
