## movidesk-cli persons

Gerencia pessoas do Movidesk (/persons)

### Synopsis

Gerencia pessoas do Movidesk. O endpoint /persons serve agentes,
clientes, empresas e departamentos — diferenciados por personType
(1=Pessoa, 2=Empresa, 4=Departamento) e profileType (1=Agente, 2=Cliente,
3=Ambos).

Filtros e projeções OData aplicam-se em list. Os valores de campos
personalizados seguem o mesmo read-merge-patch dos chamados para evitar
a armadilha "apaga entradas ausentes" do Movidesk.

### Options

```
  -h, --help   help for persons
```

### Options inherited from parent commands

```
      --compact         JSON compacto (sem indentação)
      --no-color        desativa cores na saída
      --no-retry        desativa retentativa automática em 429/5xx
  -o, --output string   formato de saída: json|table|csv (padrão: do tenant ou 'json')
      --tenant string   nome do tenant (sobrepõe o tenant atual; env: MOVIDESK_TENANT)
      --user string     id do usuário padrão (Cod. Ref.) usado em createdBy nas escritas; sobrepõe a configuração do tenant; env: MOVIDESK_USER
  -v, --verbose         log detalhado em stderr
```

### SEE ALSO

* [movidesk-cli](movidesk-cli.md)	 - CLI para a API REST do Movidesk
* [movidesk-cli persons create](movidesk-cli_persons_create.md)	 - Cria uma pessoa a partir de corpo JSON, template ou substituições --set
* [movidesk-cli persons customfields](movidesk-cli_persons_customfields.md)	 - Lê e escreve campos personalizados de pessoa (read-merge-patch)
* [movidesk-cli persons delete](movidesk-cli_persons_delete.md)	 - Exclui uma pessoa de forma permanente (DELETE /persons?id=)
* [movidesk-cli persons get](movidesk-cli_persons_get.md)	 - Obtém uma pessoa pelo id (Cod. Ref.)
* [movidesk-cli persons list](movidesk-cli_persons_list.md)	 - Lista pessoas
* [movidesk-cli persons update](movidesk-cli_persons_update.md)	 - Aplica patch em uma pessoa por id

