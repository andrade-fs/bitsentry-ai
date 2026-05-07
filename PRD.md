
# PRD.md — bitsentry-ai

## 1. Visión

`bitsentry-ai` será una CLI/TUI para instalar, configurar y orquestar un ecosistema de agentes AI orientado a investigación, escritura técnica, desarrollo guiado por especificación y flujos especializados de ciberseguridad.

El objetivo inicial no es construir todos los flujos finales, sino crear una base sólida, extensible y mantenible que permita:

- Instalar el ecosistema con un one-liner.
- Detectar el sistema operativo y dependencias disponibles.
- Detectar agentes AI instalados, empezando como mínimo por OpenCode.
- Instalar/configurar componentes opcionales como MCPs, Engram, Context7 y skills.
- Exponer una TUI limpia y fácil de entender.
- Permitir elegir perfiles de uso.
- Preparar una arquitectura para futuros flujos:
  - SDD: Spec Driven Development.
  - SDR: Spec Driven Research.
  - Red Team / Bug Bounty research workflows.
  - Blog/notes workflows para Bitsentry.

La inspiración funcional viene de herramientas tipo `gentle-ai`, pero `bitsentry-ai` tendrá un enfoque más personalizado para investigación, notas de calidad, ciberseguridad, marca personal Bitsentry y orquestación de flujos con agentes especializados.

---

## 2. Objetivo de esta fase

Crear la **Fase 1 / Bootstrap MVP** del proyecto.

Esta fase debe entregar un repositorio funcional con:

1. Estructura base del proyecto.
2. Binario/CLI ejecutable llamado `bitsentry-ai`.
3. Instalador one-liner inicial.
4. TUI básica.
5. Detección de sistema operativo.
6. Detección de dependencias mínimas.
7. Detección inicial de agentes AI instalados.
8. Soporte mínimo para OpenCode.
9. Sistema de configuración local.
10. Sistema de perfiles básico.
11. Arquitectura preparada para añadir componentes, skills y flujos más adelante.

No se debe implementar todavía toda la lógica avanzada de SDR, SDD o Red Team. En esta fase solo se deben crear los cimientos.

---

## 3. Nombre del proyecto

Nombre del proyecto:

```txt
bitsentry-ai
````

Nombre del binario:

```txt
bitsentry-ai
```

Comando esperado:

```bash
bitsentry-ai
```

Comando de instalación futuro:

```bash
curl -fsSL https://raw.githubusercontent.com/bitsentry/bitsentry-ai/main/install.sh | bash
```

Durante desarrollo local:

```bash
git clone <repo>
cd bitsentry-ai
./install.sh
bitsentry-ai
```

---

## 4. Público objetivo

El usuario principal inicial será el propio creador de Bitsentry.

Casos de uso iniciales:

* Preparar rápido un entorno AI local.
* Configurar OpenCode con memoria, MCPs y skills.
* Elegir qué componentes instalar.
* Crear perfiles según el tipo de trabajo:

  * Desarrollo.
  * Research.
  * Blog.
  * OSCP / Pentesting.
  * Bug bounty.
  * Red team.
* En el futuro, ejecutar flujos guiados paso a paso para generar notas técnicas de calidad.

---

## 5. Principios de producto

### 5.1 Simplicidad

La TUI debe ser clara, directa y sin ruido.

El usuario no debe tener que recordar comandos complejos.

### 5.2 Modularidad

Todo debe estar diseñado como componentes instalables/activables:

* Agents.
* MCPs.
* Skills.
* Memory providers.
* Workflows.
* Profiles.
* Prompts.
* Templates.

### 5.3 Instalación progresiva

El usuario podrá elegir qué instalar.

Ejemplo:

* Solo SDR.
* Solo SDD.
* Solo Engram.
* Solo OpenCode integration.
* Todo el stack Bitsentry.

### 5.4 No acoplarse a un único agente

La primera versión debe soportar OpenCode como mínimo, pero la arquitectura debe permitir añadir:

* Claude Code.
* Codex.
* Gemini CLI.
* Cursor.
* VS Code Copilot.
* Otros agentes futuros.

### 5.5 Configuración explícita

El usuario debe poder ver y modificar:

* Qué agente usa.
* Qué modelo usa cada flujo.
* Qué memoria usa.
* Qué MCPs están activos.
* Qué skills están instaladas.
* Qué perfil está activo.

---

## 6. Alcance de la Fase 1

### Incluido

La Fase 1 debe implementar:

* CLI base `bitsentry-ai`.
* TUI principal.
* Menú inicial.
* Instalador `install.sh`.
* Detección de sistema operativo.
* Detección de shell.
* Detección de arquitectura.
* Detección de dependencias.
* Detección de OpenCode.
* Configuración local en el home del usuario.
* Estructura base para componentes.
* Estructura base para perfiles.
* Estructura base para workflows.
* Logs básicos.
* Modo dry-run.
* Modo doctor/check.

### No incluido todavía

No implementar aún:

* SDR completo.
* SDD completo.
* Red Team workflows completos.
* Generación real de posts.
* Sincronización avanzada con Engram Cloud.
* Instalación completa de todos los MCPs.
* Marketplace de skills.
* Integración real con todos los agentes.
* UI compleja.
* Sistema de plugins dinámico avanzado.

---

## 7. Stack técnico propuesto

Para esta fase se prefiere una herramienta compilable, portable y fácil de distribuir.

Opción recomendada:

```txt
Go
```

Motivos:

* Binario único.
* Buena compatibilidad macOS/Linux.
* Ideal para CLI/TUI.
* Fácil distribución.
* Buena experiencia con Bubble Tea/Lip Gloss si se usa TUI avanzada.

Librerías recomendadas:

```txt
cobra          -> CLI commands
bubbletea      -> TUI
lipgloss       -> estilos TUI
bubbles        -> componentes TUI
survey/huh     -> formularios interactivos, si se prefiere algo más simple
```

La Fase 1 puede comenzar con TUI sencilla y evolucionar después.

---

## 8. Estructura inicial del repositorio

Crear una estructura similar a:

```txt
bitsentry-ai/
├── README.md
├── PRD.md
├── install.sh
├── go.mod
├── go.sum
├── cmd/
│   └── bitsentry-ai/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── app.go
│   ├── tui/
│   │   ├── model.go
│   │   ├── menu.go
│   │   └── styles.go
│   ├── system/
│   │   ├── detect.go
│   │   ├── dependencies.go
│   │   └── shell.go
│   ├── agents/
│   │   ├── agent.go
│   │   └── opencode.go
│   ├── components/
│   │   ├── component.go
│   │   ├── engram.go
│   │   ├── context7.go
│   │   └── mcp.go
│   ├── profiles/
│   │   ├── profile.go
│   │   └── defaults.go
│   ├── config/
│   │   ├── config.go
│   │   └── paths.go
│   ├── workflows/
│   │   ├── workflow.go
│   │   ├── sdd.go
│   │   ├── sdr.go
│   │   └── redteam.go
│   └── logs/
│       └── logger.go
├── assets/
│   ├── skills/
│   ├── prompts/
│   ├── profiles/
│   └── workflows/
├── scripts/
│   ├── build.sh
│   └── dev.sh
└── docs/
    ├── architecture.md
    ├── install.md
    └── roadmap.md
```

---

## 9. Configuración local

La configuración debe guardarse en:

```txt
~/.bitsentry-ai/
```

Estructura esperada:

```txt
~/.bitsentry-ai/
├── config.yaml
├── profiles/
│   ├── default.yaml
│   ├── research.yaml
│   ├── dev.yaml
│   └── redteam.yaml
├── logs/
│   └── bitsentry-ai.log
├── cache/
└── state.json
```

Ejemplo inicial de `config.yaml`:

```yaml
version: 1

active_profile: default

agents:
  opencode:
    enabled: true
    detected: false
    path: null

components:
  engram:
    enabled: false
    installed: false
  context7:
    enabled: false
    installed: false
  mcps:
    enabled: false
    installed: []

workflows:
  sdd:
    enabled: false
  sdr:
    enabled: false
  redteam:
    enabled: false
```

---

## 10. TUI inicial

La TUI debe tener un menú principal claro.

Primera versión:

```txt
bitsentry-ai

? What do you want to do?

  1. Install / Setup
  2. System check
  3. Detect AI agents
  4. Components
  5. Profiles
  6. Workflows
  7. Settings
  8. Exit
```

---

## 11. Menú: Install / Setup

Este será el flujo principal inicial.

Debe hacer:

1. Detectar sistema operativo.
2. Detectar arquitectura.
3. Detectar shell.
4. Detectar dependencias.
5. Detectar agentes AI.
6. Mostrar resumen.
7. Preguntar qué se desea instalar/configurar.

Ejemplo:

```txt
System detected:

OS: macOS
Arch: arm64
Shell: zsh

Dependencies:
✓ git
✓ curl
✓ go
✗ node
✗ pnpm

AI Agents:
✓ opencode found at /usr/local/bin/opencode
✗ claude not found
✗ codex not found
✗ gemini not found

What do you want to configure?

[x] OpenCode
[x] Engram
[x] Context7
[ ] Red Team workflows
[ ] SDR workflows
[ ] SDD workflows
```

En la Fase 1 no hace falta que instale todo realmente. Puede dejar stubs o mensajes claros tipo:

```txt
Component installation not implemented yet.
Prepared config entry for future installation.
```

Pero OpenCode sí debe detectarse como mínimo.

---

## 12. Detección de sistema

Implementar detección de:

* Sistema operativo:

  * macOS.
  * Linux.
  * Windows WSL si aplica.
* Arquitectura:

  * amd64.
  * arm64.
* Shell:

  * zsh.
  * bash.
  * fish.
* Gestores de paquetes:

  * brew.
  * apt.
  * pacman.
  * dnf.
  * unknown.

Comando CLI futuro:

```bash
bitsentry-ai doctor
```

Debe imprimir un resumen de entorno.

---

## 13. Detección de dependencias

Detectar como mínimo:

```txt
git
curl
go
node
npm
pnpm
yarn
opencode
```

La detección debe hacerse buscando binarios en PATH.

Ejemplo:

```bash
which opencode
```

En Go:

```go
exec.LookPath("opencode")
```

---

## 14. Detección de agentes AI

Crear una interfaz común:

```go
type AgentDetector interface {
    ID() string
    Name() string
    Detect() (*AgentDetectionResult, error)
}
```

Resultado:

```go
type AgentDetectionResult struct {
    ID        string
    Name      string
    Detected  bool
    Path      string
    Version   string
    ConfigDir string
}
```

Primera implementación obligatoria:

```txt
OpenCode
```

Detectar:

* Si existe el comando `opencode`.
* Ruta del binario.
* Versión si es posible.
* Posibles rutas de configuración.

---

## 15. Componentes

Crear abstracción para componentes instalables.

Ejemplos futuros:

* Engram.
* Context7.
* MCPs.
* Skills.
* Personas.
* Prompts.
* Permissions.
* Workflows.

Interfaz inicial:

```go
type Component interface {
    ID() string
    Name() string
    Description() string
    IsInstalled() bool
    Install() error
    Uninstall() error
    Status() ComponentStatus
}
```

En Fase 1, los componentes pueden estar en modo stub, pero la arquitectura debe quedar preparada.

---

## 16. Perfiles

Un perfil define cómo se comporta un flujo o agente.

Ejemplos:

```txt
default
dev
research
blog
oscp
bug-bounty
redteam
```

Cada perfil debe poder definir:

* Agente principal.
* Modelo preferido.
* Nivel de razonamiento.
* Memory provider.
* Skills activas.
* MCPs activos.
* Workflows activos.

Ejemplo de perfil:

```yaml
id: research
name: Research
description: Profile for technical research and high-quality notes.

agent:
  default: opencode

models:
  planner: gpt-5.5-thinking
  researcher: gpt-5.5-thinking
  writer: gpt-5.5
  validator: gpt-5.5-thinking

memory:
  provider: engram
  project: bitsentry

components:
  context7: true
  engram: true

workflows:
  sdr: true
  sdd: false
  redteam: false
```

En Fase 1 solo se requiere:

* Crear perfiles default.
* Listarlos.
* Cambiar perfil activo.
* Guardar perfil activo en config.

---

## 17. Workflows futuros

Los workflows deben modelarse desde el principio, aunque no se implementen completamente.

### 17.1 SDD — Spec Driven Development

Flujo orientado a desarrollo por especificación.

Objetivo futuro:

1. Recibir idea o issue.
2. Crear spec.
3. Crear plan técnico.
4. Dividir tareas.
5. Ejecutar con agente.
6. Validar cambios.
7. Generar resumen.

### 17.2 SDR — Spec Driven Research

Flujo orientado a research y notas de calidad.

Objetivo futuro:

1. Recibir una idea.
2. Evaluar potencial.
3. Contrastar contra notas existentes.
4. Buscar ángulo diferencial.
5. Crear estructura.
6. Redactar nota.
7. Validar calidad.
8. Preparar publicación.

### 17.3 Red Team / Bug Bounty

Flujo orientado a investigación ofensiva controlada y documentación.

Objetivo futuro:

1. Definir objetivo autorizado.
2. Crear scope.
3. Enumerar hipótesis.
4. Generar checklist.
5. Registrar hallazgos.
6. Convertir hallazgos en notas.
7. Preparar writeup técnico.

Importante:

En esta fase no ejecutar ataques, exploits ni automatizaciones ofensivas. Solo preparar la arquitectura y nombres de workflows.

---

## 18. Comandos CLI esperados

Además de la TUI interactiva, crear comandos básicos:

```bash
bitsentry-ai
bitsentry-ai doctor
bitsentry-ai agents
bitsentry-ai profiles
bitsentry-ai profile use <name>
bitsentry-ai components
bitsentry-ai config path
bitsentry-ai version
```

### 18.1 `bitsentry-ai`

Abre la TUI.

### 18.2 `bitsentry-ai doctor`

Muestra diagnóstico del sistema.

### 18.3 `bitsentry-ai agents`

Lista agentes detectados.

### 18.4 `bitsentry-ai profiles`

Lista perfiles disponibles.

### 18.5 `bitsentry-ai profile use <name>`

Cambia el perfil activo.

### 18.6 `bitsentry-ai components`

Lista componentes conocidos.

### 18.7 `bitsentry-ai config path`

Muestra ruta de configuración.

### 18.8 `bitsentry-ai version`

Muestra versión del binario.

---

## 19. Instalador one-liner

Crear `install.sh`.

Responsabilidades:

1. Detectar sistema operativo.
2. Detectar arquitectura.
3. Crear directorio temporal.
4. Descargar o compilar binario.
5. Instalar en una ruta disponible.
6. Crear configuración inicial.
7. Mostrar siguiente comando.

Durante desarrollo, puede compilar localmente:

```bash
go build -o bitsentry-ai ./cmd/bitsentry-ai
```

Instalación local propuesta:

```bash
~/.local/bin/bitsentry-ai
```

Si `~/.local/bin` no existe:

```bash
mkdir -p ~/.local/bin
```

Si no está en PATH, mostrar aviso:

```txt
Add this to your shell config:

export PATH="$HOME/.local/bin:$PATH"
```

---

## 20. UX esperada

La herramienta debe hablar de forma clara.

Evitar mensajes técnicos innecesarios.

Ejemplo bueno:

```txt
✓ OpenCode detected
✓ Config directory ready
✗ Engram not installed

Next step:
Run "bitsentry-ai" and choose Install / Setup.
```

Ejemplo malo:

```txt
panic: exec: no such file or directory
```

Los errores deben ser amigables y accionables.

---

## 21. Logs

Guardar logs en:

```txt
~/.bitsentry-ai/logs/bitsentry-ai.log
```

El usuario debe poder verlos con:

```bash
bitsentry-ai logs
```

En Fase 1, si no se implementa comando `logs`, al menos crear estructura para logs.

---

## 22. Seguridad

No guardar secretos en texto plano si no es necesario.

No pedir tokens en Fase 1 salvo que sea imprescindible.

No ejecutar comandos destructivos.

No modificar configuraciones existentes sin mostrar resumen.

Antes de sobrescribir un archivo de configuración:

* Crear backup.
* Preguntar confirmación.
* O usar modo dry-run.

---

## 23. Modo dry-run

Añadir flag global:

```bash
bitsentry-ai --dry-run
```

O:

```bash
bitsentry-ai install --dry-run
```

En Fase 1, el dry-run debe mostrar qué haría sin modificar archivos.

---

## 24. Criterios de aceptación

La Fase 1 se considera completa cuando:

* El repo compila correctamente.
* `go build ./...` funciona.
* `bitsentry-ai` abre una TUI.
* `bitsentry-ai doctor` muestra sistema, shell y dependencias.
* `bitsentry-ai agents` detecta OpenCode si está instalado.
* `bitsentry-ai profiles` lista perfiles iniciales.
* `bitsentry-ai profile use research` cambia el perfil activo.
* Se crea `~/.bitsentry-ai/config.yaml`.
* `install.sh` instala el binario localmente.
* El README explica instalación y uso básico.
* La arquitectura permite añadir Engram, Context7, MCPs, skills y workflows sin reescribir el core.

---

## 25. Roadmap futuro

### Fase 1 — Bootstrap MVP

* Repo base.
* CLI.
* TUI.
* One-liner.
* Doctor.
* Detección de OpenCode.
* Config local.
* Perfiles básicos.

### Fase 2 — Components MVP

* Instalación real de Engram.
* Instalación/configuración de Context7.
* Gestión inicial de MCPs.
* Gestión inicial de skills.
* Presets de instalación.

### Fase 3 — SDR MVP

* Flujo de Spec Driven Research.
* Agentes:

  * Research planner.
  * Source evaluator.
  * Angle finder.
  * Note writer.
  * Quality validator.
* Salida en Markdown compatible con Obsidian/blog Bitsentry.

### Fase 4 — SDD MVP

* Flujo de Spec Driven Development.
* Specs.
* Plan técnico.
* Tareas.
* Validación.

### Fase 5 — Red Team / Bug Bounty workflows

* Scope-first workflows.
* Checklist.
* Research autorizado.
* Notes/writeups.
* Hallazgos.
* Templates para disclosure.

### Fase 6 — Profiles avanzados

* Modelos por agente.
* Reasoning por agente.
* Presupuestos de tokens.
* Preferencias por proyecto.
* Integración con Engram Cloud.

---

## 26. Prioridad inmediata para implementación

Implementar en este orden:

1. Crear estructura del repo.
2. Inicializar Go module.
3. Crear comando `bitsentry-ai version`.
4. Crear comando `bitsentry-ai doctor`.
5. Crear detección de dependencias.
6. Crear detección de OpenCode.
7. Crear config local.
8. Crear perfiles iniciales.
9. Crear TUI básica.
10. Crear `install.sh`.
11. Crear README inicial.
12. Añadir stubs para components/workflows.

---

## 27. Resultado esperado para esta iteración

Al terminar esta iteración, se debe poder hacer:

```bash
git clone <repo>
cd bitsentry-ai
./install.sh
bitsentry-ai doctor
bitsentry-ai agents
bitsentry-ai
```

Y obtener una TUI mínima funcional desde la que se puedan ver:

* Sistema detectado.
* Dependencias.
* Agentes.
* Perfiles.
* Componentes futuros.
* Workflows futuros.

---

## 28. Nota de diseño

No construir una herramienta monolítica.

`bitsentry-ai` debe ser un orquestador/configurador modular.

La lógica de instalación, detección, perfiles, componentes y workflows debe estar separada para que en el futuro sea sencillo añadir:

* Nuevos agentes.
* Nuevos MCPs.
* Nuevas skills.
* Nuevos flujos.
* Nuevos perfiles.
* Nuevos presets.

---

## 29. Definición corta

`bitsentry-ai` es el bootstrapper y orquestador AI de Bitsentry.

Su primera misión es preparar el entorno.

Su misión futura es ejecutar flujos de investigación, desarrollo y documentación técnica de alta calidad mediante agentes especializados.

````

Para pasárselo a `opencode`, yo usaría este prompt justo después de crear el `PRD.md`:

```txt
Read PRD.md and implement only Phase 1.

Do not implement SDR, SDD or Red Team workflows yet.
Create the Go repository structure, CLI commands, system detection, OpenCode detection, config handling, basic profiles, minimal TUI and install.sh.

Keep the architecture modular and extensible.
Prioritize clean code, small packages and simple UX.
After implementation, show me:
1. files created
2. commands to build
3. commands to test manually
4. what is still stubbed
````

Mi recomendación: en esta primera fase no metas aún instalación real de Engram/Context7/MCPs. Primero consigue que `bitsentry-ai doctor`, `bitsentry-ai agents`, `bitsentry-ai profiles` y la TUI funcionen perfecto. Luego ya abrimos la Fase 2.

[1]: https://github.com/Gentleman-Programming/gentle-ai?utm_source=chatgpt.com "Gentleman-Programming/gentle-ai"
