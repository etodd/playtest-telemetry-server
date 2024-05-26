<!DOCTYPE html>
<html>
<head>
    <title>Playtest Admin</title>
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
    <script>
        function req(path, body, callback) {
            fetch(path, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(body),
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                })
                .then(data => {
                    if (callback) {
                        callback();
                    }
                })
                .catch(error => {
                    alert('Error: ' + error);
                });
        }
    </script>
</head>
<body>
    <h1>Playtest Admin</h1>
    <table>
    {{- if .Versions }}
        <tr>
            <th>Version</th>
            <th>Files</th>
            <td></td>
        </tr>
        {{- range $_, $version := .Versions }}
            <tr>
                <td>{{ $version.Version }}</td>
                <td><a href="telemetry/download?version={{ $version.Version }}">{{ $version.FileCount }} files</td>
                <td><a href="#{{ $version.Version }}" onclick="req('telemetry/clear', { version: '{{ $version.Version }}' }, function() { location.reload(); })">Clear</td>
            </tr>
        {{- end }}
    {{- else }}
        <tr>
            <td colspan="3">No telemetry uploaded yet</td>
        </tr>
    {{- end }}
    </table>
</body>
</html>