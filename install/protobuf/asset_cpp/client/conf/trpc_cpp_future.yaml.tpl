{{- $app := .AppName -}} {{- if eq (len .AppName) 0 -}} {{- $app = "appdemo" -}} {{- end -}}
{{- $server := .ServerName -}} {{- if eq (len .ServerName) 0 -}}  {{- $server = "serverdemo" -}} {{- end -}}
global:
  threadmodel:
    default:
      - instance_name: default_instance
        io_handle_type: separate
        io_thread_num: 2
        handle_thread_num: 6

client:
  service:
    {{- range $idx, $svc := .Services }}
    - name: trpc.{{ $app }}.{{ $server }}.{{ $svc.Name }}
      target: 127.0.0.1:5432{{ $idx }}
      protocol: trpc
      network: tcp
      selector_name: direct
    {{- end }}

plugins:
  log:
    default:
      - name: default
        sinks:
          local_file:
            filename: trpc_client.log