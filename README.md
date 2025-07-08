

# Bounty Analyzer

Ferramenta automatizada para análise e reconhecimento em programas de Bug Bounty. Desenvolvida em Go para máxima performance e fácil integração em pipelines de pentest/recon.

## 🚀 Funcionalidades

- Reconhecimento automatizado de alvos
- Coleta e análise de subdomínios, endpoints e parâmetros
- Identificação de potenciais vetores (IDOR, SSRF, etc)
- Integração com APIs ou wordlists customizadas
- Geração de relatórios ou integração com ferramentas externas

## 📦 Instalação

```bash
git clone https://github.com/404xploit/bounty-analyzer.git
cd bounty-analyzer
go build
# ou
go install ./...
```

## ⚡️ Uso Rápido

```bash
./bounty-analyzer -d target.com -o resultado.txt
```

Parâmetros comuns:

- -d: domínio ou lista de domínios
- -o: arquivo de saída
- --threads: customizar paralelismo
- --mode: modo de análise (recon, bruteforce, endpoints, etc)

## 🔥 Exemplos de Comando

Recon básico:
```bash
./bounty-analyzer -d lista.txt --mode recon
```

Exploração de parâmetros:
```bash
./bounty-analyzer -d alvo.com --mode params
```

## 🧬 Customização

- Adapte wordlists em /lists/
- Integre com scripts externos via STDIN/STDOUT
- Edite módulos em Go pra incluir payloads próprios ou lógicas de bypass

## 🤝 Contribuição

Pull requests, issues e sugestões são bem-vindos.
Se for adaptar pra cenários underground, só cite as novas técnicas no PR.

## ⚠️ Aviso

Uso destinado a pesquisas de segurança e pentest autorizado. Responsabilidade total do usuário.

