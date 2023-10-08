{{- $app := .AppName -}} {{- if eq (len .AppName) 0 -}} {{- $app = "appdemo" -}} {{- end -}}
{{- $server := .ServerName -}} {{- if eq (len .ServerName) 0 -}}  {{- $server = "serverdemo" -}} {{- end -}}
global:
  threadmodel:
    default:
      - instance_name: default_instance
        io_handle_type: separate
        io_thread_num: 2
        handle_thread_num: 6

server:
  app: {{ $app }}
  server: {{ $server }}
  admin_port: 6666
  admin_ip: 0.0.0.0
  service:
    {{- range $idx, $svc := .Services }}
    - name: trpc.{{ $app }}.{{ $server }}.{{ $svc.Name }}
      protocol: trpc
      network: tcp
      ip: 0.0.0.0
      port: 5432{{ $idx }}
    {{- end }}

plugins:
  log:
    default:
      - name: default
        sinks:
          local_file:
            filename: trpc_server.log