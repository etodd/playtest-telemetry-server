<!DOCTYPE html>
<html>
<head>
    <title>Playtest Telemetry</title>
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
    <h1>Playtest Telemetry</h1>
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
                <td><a href="/download?version={{ $version.Version }}">{{ $version.FileCount }} files</td>
                <td><a href="#" onclick="req('/clear', { version: '{{ $version.Version }}' }, function() { location.reload(); })">Clear</td>
            </tr>
        {{- end }}
    {{- else }}
        <tr>
            <td colspan="3">No files uploaded yet</td>
        </tr>
    {{- end }}
    </table>
</body>
</html>