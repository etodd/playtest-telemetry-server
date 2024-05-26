<!DOCTYPE html>
<html>
<head>
    <title>Playtest Builds</title>
    <style>
    html {
        font-family: sans-serif;
    }
    @media (prefers-color-scheme: dark) {
        body {
            background-color: #333;
            color: #fff;
        }
        a {
            color: #adf;
        }
        a:visited {
            color: #faf;
        }
    }
    </style>
</head>
<body>
    <h1>Playtest Builds</h1>
    <table>
    {{- if .Builds }}
        {{- range $_, $filename := .Builds }}
        <tr>
            <td><a href="/builds/{{ $filename }}">{{ $filename }}</a></td>
        </tr>
        {{- end }}
    {{- else }}
        <tr>
            <td>No builds available yet</td>
        </tr>
    {{- end }}
    </table>
</body>
</html>