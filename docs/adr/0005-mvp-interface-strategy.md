# ADR 0005: Interface Inicial do MVP

## Contexto

O OrchestraOS precisa de uma forma inicial de operar o Orchestrator, criar tasks, acompanhar runs, aprovar ferramentas e inspecionar evidencias. As opcoes consideradas sao scripts soltos, CLI, Desktop e Web.

O projeto ainda esta em fase de validacao do nucleo: task graph, prompt composer, sandbox, agentes, eventos, politicas e auditoria. Uma interface rica antes desse nucleo estar validado pode aumentar custo sem reduzir o maior risco tecnico.

GitHub sera o complemento operacional principal da CLI. Conectores de chat podem existir no futuro, mas nao fazem parte do caminho critico do MVP.

## Decisão

O MVP tera uma progressao em tres camadas:

1. **Scripts de bootstrap** para desenvolvimento local e validacao dos componentes internos.
2. **CLI fina** como primeira interface oficial do MVP para criar tasks, iniciar runs, acompanhar eventos, enviar mensagens, aprovar ou negar ferramentas e revisar evidencias.
3. **GitHub** como superficie externa principal para issues, pull requests, revisoes, checks e evidencias.

Desktop e Web nao entram como requisito do primeiro MVP.

Desktop permanece como candidato forte para uma experiencia local rica no futuro, especialmente para uso solo, visualizacao de traces e intervencao em agentes. Web permanece como candidato quando houver servidor remoto, multiplos usuarios, acesso de equipe ou necessidade de painel operacional compartilhado.

## Consequências

- O nucleo do Orchestrator pode ser validado sem depender de UI complexa.
- Scripts continuam permitidos, mas nao devem virar interface permanente ou contrato publico.
- A CLI cria um contrato operacional testavel e automatizavel.
- GitHub complementa a CLI com revisao, historico e integracao via pull requests.
- Chat fica opcional e nao substitui comandos auditaveis, PRs ou estado persistido.
- A decisao Desktop vs Web sera revisitada depois que o fluxo local fim a fim estiver funcionando.

## Alternativas consideradas

- **Apenas scripts**: rapido para prototipo, mas fraco como contrato operacional e dificil de padronizar.
- **Desktop desde o inicio**: atraente para experiencia local, mas adiciona custo antes de validar o control plane.
- **Web desde o inicio**: bom para colaboracao futura, mas prematuro para um MVP local-first.
- **Chat como interface principal**: pratico para conversa, mas limitado para auditoria, operacao detalhada, replay e seguranca.
