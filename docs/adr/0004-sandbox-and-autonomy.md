# ADR 0004: Sandbox e Autonomia Inicial

## Contexto

Agentes executarao tasks com autonomia concedida pelo usuario. Eles poderao editar arquivos, rodar comandos e interagir com GitHub conforme politica. No futuro, poderao interagir com chat e outros sistemas quando houver conectores aprovados.

O projeto precisa proteger contra erro acidental e reduzir risco de codigo potencialmente malicioso. Tambem precisa respeitar a politica de autonomia progressiva definida em `AGENTS.md`.

## Decisão

No MVP, o nivel de autonomia aprovado sera **Nivel 2: IA implementa com revisao humana**.

O sandbox inicial sera composto por:

- Um git worktree por task.
- Uma branch por task.
- Um container por agente.
- Montagem apenas do worktree e caminhos explicitamente aprovados.
- Usuario nao-root quando possivel.
- Sem container privilegiado.
- Sem Docker socket montado.
- Sem home do usuario montada.
- Limites de CPU, memoria, processos e tempo.
- Rede bloqueada por padrao ou liberada por politica.
- Segredos injetados apenas por aprovacao explicita.
- Eventos e logs persistidos para auditoria.

Acoes de maior risco exigem aprovacao explicita, incluindo rede, dependencias externas, segredos, escrita fora do worktree, push, PR, comandos destrutivos e mudancas de politica.

Niveis 4 e 5 nao estao aprovados para o MVP.

## Consequências

- O MVP pode validar execucao real com risco controlado.
- Agentes conseguem implementar, mas nao podem operar livremente fora do escopo aprovado.
- O sistema precisa implementar checkpoints e pedidos de aprovacao.
- Docker nao deve ser considerado isolamento suficiente contra adversarios fortes.
- A evolucao para gVisor ou Firecracker deve ser planejada antes de executar codigo de origem pouco confiavel.

## Alternativas Consideradas

- **Sem sandbox, apenas worktree**: simples, mas arriscado para comandos, segredos e sistema de arquivos.
- **VM forte desde o inicio**: mais seguro, mas aumenta custo e complexidade antes do MVP.
- **Autonomia nivel 3 ou maior por padrao**: acelera fluxo, mas viola a progressao definida no projeto.
- **Permissoes livres com logs posteriores**: registra danos, mas nao previne acoes de alto risco.
