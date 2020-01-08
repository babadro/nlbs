### API Run
.\api.exe -port=8888 -dbuser=postgres -dbname=test -dbpass=postgrespass -clean=true

### Postprocessing Run
.\postprocess.exe -dbuser=postgres -dbname=test -dbpass=postgrespass -timeout=1s
