# Sandbox e Autonomia

## Objetivo

Proteger o projeto contra erro acidental e reduzir risco de execucao de codigo potencialmente malicioso.

No MVP local, Docker + git worktree oferecem isolamento pratico, mas nao devem ser tratados como fronteira perfeita de seguranca. Quando o risco aumentar, o sandbox deve evoluir para gVisor ou Firecracker.

## Camadas de Isolamento

### Worktree

Cada task executa em um worktree proprio e branch propria. Isso evita conflitos diretos entre agentes e permite revisar, descartar ou integrar mudancas de forma independente.

### Container

Cada agente executa em um container separado. Regras iniciais:

- Nao usar container privilegiado.
- Nao montar Docker socket.
- Nao montar home do usuario.
- Montar apenas o worktree da task e diretorios explicitamente aprovados.
- Usar usuario nao-root quando possivel.
- Aplicar limite de CPU, memoria, processos e tempo de execucao.
- Bloquear ou limitar rede por padrao.
- Injetar segredos apenas por aprovacao e com escopo minimo.
- Registrar comandos, saidas relevantes e exit codes.

### Rede

Rede deve ser negada por padrao para tasks de baixo risco que nao precisam de internet. Quando necessaria, deve ser liberada por politica ou aprovacao, preferencialmente com allowlist.

### Segredos

Segredos nao devem ser copiados para worktrees, logs, prompts ou artefatos. O Orchestrator deve controlar injecao temporaria de credenciais e registrar qual politica permitiu o acesso.

### Ferramentas

Toda ferramenta deve ter classificacao de risco. Ferramentas de leitura e validacao podem ser autoaprovadas em escopos seguros. Ferramentas destrutivas, externas ou com segredos exigem aprovacao.

## Politica Inicial de Autonomia

Nivel aprovado para o MVP:

- **Nivel 2**: IA implementa em sandbox com revisao humana.

Permitido por padrao no MVP:

- Ler arquivos do worktree da task.
- Consultar docs versionadas do projeto.
- Editar arquivos dentro do worktree da task.
- Rodar validacoes locais sem rede quando disponiveis.
- Produzir diff, resumo e evidencias.

Exige aprovacao explicita:

- Acessar rede.
- Instalar dependencias externas.
- Ler ou usar segredos.
- Escrever fora do worktree da task.
- Fazer push para remoto.
- Abrir PR.
- Executar comandos destrutivos.
- Alterar politicas de autonomia.
- Interagir com sistemas externos alem do escopo da task.

Nao aprovado no MVP:

- **Nivel 4**: executar tarefas operacionais de baixo risco sem revisao.
- **Nivel 5**: operar dominios aprovados com autonomia ampla.

Nivel 3 pode ser habilitado por task ou ADR especifica quando o sistema ja conseguir abrir PRs com testes e evidencias, sempre com trilha de auditoria.

## Checkpoints

Agentes devem parar em checkpoints seguros para:

- Ler mensagens pendentes do Orchestrator.
- Confirmar mudanca de escopo.
- Solicitar aprovacao de ferramenta.
- Registrar progresso.
- Validar se a task ainda esta dentro da politica permitida.

## Encerramento e Limpeza

Ao finalizar uma task, o Orchestrator deve:

- Coletar diff, logs e evidencias.
- Registrar validacoes.
- Encerrar o processo do agente.
- Parar e remover container.
- Manter ou remover worktree conforme politica de retencao.
- Enviar status à CLI/GitHub.
