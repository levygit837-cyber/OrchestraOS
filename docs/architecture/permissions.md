# Permissões e Ferramentas

## Objetivo

Definir quais ferramentas agentes podem usar automaticamente, quais exigem aprovacao e quais ficam bloqueadas no MVP.

## Classes de Risco

| Classe | Significado |
| --- | --- |
| `safe` | Baixo risco dentro do worktree e sem rede. |
| `guarded` | Permitido apenas com limites claros ou ownership. |
| `approval_required` | Exige decisao humana ou politica explicita. |
| `destructive` | Pode apagar, sobrescrever, publicar ou vazar informacao. |
| `forbidden` | Nao permitido no MVP. |

## Ferramentas Seguras Para Agentes Paralelos

| Ação | Política MVP | Condições |
| --- | --- | --- |
| Ler arquivos do worktree | Autoaprovada | Apenas caminhos da task. |
| Buscar texto com `rg` | Autoaprovada | Apenas worktree e docs aprovadas. |
| Consultar `git status` e `git diff` | Autoaprovada | Apenas branch/worktree da task. |
| Editar arquivos sob ownership | Autoaprovada | Dentro do worktree e escopo da work unit. |
| Criar arquivos de teste no escopo | Autoaprovada | Dentro de caminhos aprovados. |
| Rodar testes locais sem rede | Autoaprovada | Com timeout e limite de recursos. |
| Rodar lint/format local | Autoaprovada | Sem instalar dependencias novas. |
| Atualizar Agent Task Ledger | Autoaprovada | Apenas ledger da propria work unit. |
| Emitir eventos e logs | Autoaprovada | Sem segredos. |
| Criar artefatos de evidencia | Autoaprovada | Dentro do diretorio de artefatos da run. |
| Solicitar ferramenta ausente | Autoaprovada | Pedido estruturado, sem executar a ferramenta. |

## Ações Que Exigem Aprovação

| Ação | Motivo |
| --- | --- |
| Acessar rede | Pode vazar contexto, baixar codigo ou depender de servico externo. |
| Instalar dependencia | Altera supply chain e reprodutibilidade. |
| Ler segredo | Exige escopo minimo e justificativa. |
| Usar token de GitHub ou conector externo | Interage com sistemas externos. |
| Fazer push | Publica alteracoes fora do sandbox local. |
| Abrir PR | Cria efeito externo no GitHub. |
| Enviar mensagem por chat/conector externo | Comunica em nome do sistema ou usuario. |
| Escrever fora do worktree | Pode afetar repositorio real ou maquina host. |
| Alterar politica de autonomia | Muda limites de seguranca. |
| Executar comando destrutivo | Pode apagar estado ou evidencias. |
| Rodar migration em banco compartilhado | Pode alterar estado persistente. |
| Acessar arquivos fora do escopo | Pode vazar contexto entre tasks. |
| Expandir toolset da sessão | Pode ampliar capacidade operacional do agente. |
| Reconfigurar prompt/toolset de sessão | Pode alterar comportamento e permissões durante a run. |

## Ações Bloqueadas no MVP

| Ação | Motivo |
| --- | --- |
| Montar Docker socket | Permite escapar para o host. |
| Container privilegiado | Enfraquece isolamento. |
| Montar home do usuario | Exposicao de segredos e arquivos pessoais. |
| Usar `sudo` dentro do sandbox | Escala privilegio. |
| Apagar worktree de outra task | Viola isolamento. |
| Alterar historico remoto com force push | Alto risco operacional. |
| Executar scripts remotos via pipe | Supply chain e auditoria fracas. |
| Usar segredos em prompt ou log | Risco de vazamento. |

## Política de Comandos Shell

Shell deve ser tratado como ferramenta sensivel.

Autoaprovado quando:

- comando e somente leitura ou validacao local;
- roda dentro do worktree;
- nao usa rede;
- nao escreve fora do escopo;
- tem timeout;
- nao contem redirecionamento destrutivo.

Exige aprovacao quando:

- instala pacote;
- usa rede;
- apaga arquivos;
- altera permissoes;
- modifica git remoto;
- executa binario desconhecido;
- toca em banco persistente.

## Ownership e Conflitos

Antes de iniciar agentes paralelos, o Orchestrator deve atribuir:

- `owned_paths`: caminhos que a work unit pode editar;
- `read_paths`: caminhos que pode ler;
- `blocked_paths`: caminhos proibidos;
- `shared_paths`: caminhos que exigem lock ou execucao serial.

Se duas work units precisam editar o mesmo arquivo, o Orchestrator deve:

1. serializar a execucao;
2. dividir melhor a task;
3. ou exigir aprovacao explicita para concorrencia.

## Decisões de Ferramenta

Toda decisao deve registrar:

- ferramenta;
- input;
- motivo do agente;
- classificacao de risco;
- politica aplicada;
- decisor;
- resultado;
- timestamp.

## Toolset Por AgentSession

Cada `AgentSession` deve receber um toolset minimo, definido pelo Orchestrator com base em:

- perfil do agente;
- objetivo da work unit;
- ownership de arquivos;
- risco da task;
- validacoes esperadas;
- nivel de autonomia aprovado.

O agente nao pode adicionar ferramentas por conta propria. Ele pode solicitar expansao do toolset. A decisao pertence ao Orchestrator e deve virar evento.

Quando a expansao mudar significativamente o comportamento da sessao, o Orchestrator deve criar nova sessao ou reconfiguracao auditavel em vez de alterar ferramentas silenciosamente.
