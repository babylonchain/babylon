<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.18.1/swagger-ui.css">
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@4.18.1/favicon-16x16.png">
</head>
<body>
    <div id="swagger-ui"></div>

    <script src="https://unpkg.com/swagger-ui-dist@4.18.1/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: {{ .URL }},
                dom_id: "#swagger-ui",
                deepLinking: true,
                layout: "BaseLayout",
            });
        }
    </script>
</body>
</html>
