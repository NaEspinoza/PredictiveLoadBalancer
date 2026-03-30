# Predictive Sentinel

Balanceador P2C + EWMA optimizado para baja latencia y alta disponibilidad.

## Características
- Algoritmo: Power of Two Choices (P2C) + EWMA de latencia.
- Penalización por errores 5xx y circuit-breaker básico.
- Observabilidad con Prometheus (/metrics).
- Healthchecks activos y select P2C que evita backends marcados como `alive=false`.
- Contenedorizado y orquestación con docker-compose para pruebas.

## Requisitos
- Docker & Docker Compose
- Go 1.20+
- Python 3.9+ (para Locust)

## Quickstart (Docker Compose)

1. Copia variables de ejemplo:
   ```bash
   cp .env.example .env
   ```

2. Levanta los servicios:

   ```bash
   docker-compose up --build
   ```
3. Predictive Sentinel estará en `http://localhost:8080`.

   * Métricas Prometheus: `http://localhost:8080/metrics` 
   * Backends de ejemplo: `http://localhost:9001`, `9002`, `9003` 
   * Prometheus UI: `http://localhost:9090` (si se expone)

## Pruebas de carga

* `tests/locustfile.py` contiene el escenario básico. Instala Locust y abre `http://localhost:8089`.

## Estructura

Ver `./` para la lista completa de archivos.

## Contribuir

* Fork & PR. Añade tests y actualiza `README`.

## Licencia

MIT
