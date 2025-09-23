# Build Metadata & Deep Linking

Adds commit/version/time resource attributes and span events containing a ready-to-use Pyroscope query URL when trace/profile correlation is enabled.

See generator-otel and loader-otel code for `pyroscope.link` span event creation.

Steps:
1. Build with ldflags injecting git commit, version, build time.
2. Run with `PYROSCOPE_ENABLE=true PYROSCOPE_TRACE_CORRELATION=true`.
3. Open trace in Tempo → locate `pyroscope.link` event.
4. Follow `pyroscope.query_url` to filtered profile view.
