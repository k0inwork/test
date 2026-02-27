# JSON Dynamic API Middleware

The PUM system features a highly advanced `json_middleware` that allows the entire application to act as both a traditional multi-page web application and a REST-like JSON API without requiring duplicate view logic.

## 1. Automated API Generation

The middleware (in `products_app/middleware/json.py`) dynamically transforms template-rendering views into JSON-returning views.

- **Trigger:** Any request with the query parameter `?json=true` will be handled by the JSON engine.
- **Monkey-patching:** Upon the first such request, the middleware scans all registered views in the project and patches their `get` and `post` methods, as well as the standard Django `loader.render_to_string` function.

## 2. Context Interception and Serialization

When a view is called with `json=true`:
1. **Context Capture:** The middleware intercepts the "Context" dictionary that the view would normally pass to its HTML template.
2. **Form Error Extraction:** If the view contains a Django form (e.g., a login or configuration form), the middleware automatically extracts any validation errors.
3. **Recursive Serialization:** It passes the context and errors to the `jsonify` engine, converting complex database objects, querysets, and timestamps into a clean JSON structure.
4. **Redirect Mapping:** HTTP redirects (301/302) are converted into a JSON response: `{"url_redirect": "/new/path/"}`.

## 3. Client-Side Filtering

A unique feature of this middleware is the ability for the client to request only specific parts of the context to reduce bandwidth.

- **Usage:** By adding `json_<key>` parameters to the URL, the client can filter the output.
- **Example:** `GET /products/1/?json=true&json_name&json_state`
  - The middleware will intercept the full product context but only return the `name` and `state` fields in the final JSON response.

## 4. Integration with Performance Layer

- **Caching:** The middleware is aware of the `json_cached` decorator. If a view has a pre-computed JSON version in Redis, the middleware will serve it directly, bypassing the view execution entirely.
- **Async Support:** The middleware is fully compatible with Django's asynchronous views, providing an `async_newf` wrapper that correctly handles `await` calls during context generation.

## 5. Developer Benefit

This architecture allows developers to write standard Django Class-Based Views (CBVs) focused on data and business logic. The `json_middleware` handles the complexity of exposing that same logic to mobile apps, external scripts, or modern frontend frameworks automatically.
