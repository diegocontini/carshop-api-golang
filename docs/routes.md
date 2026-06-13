# Rotas da API (Go)

Este documento mostra cada rota da API com o formato esperado de entrada e saída.

Espelha `../../CarShopApi/src/Docs/routes.md`. Diferenças intencionais introduzidas pelo port para Go:

- `GET /` redireciona para `/docs/routes.md` em vez de `/swagger` (não há Swagger no build Go).
- `GET /api/v1/user` e `GET /api/v1/user/{id}` **omitem** o campo `password` na resposta. POST e PUT continuam aceitando `password`, que é armazenado com bcrypt. Ver `../bug-fixes-report/004-plaintext-passwords.md`.
- `POST /api/v1/user` retorna `409 Conflict` (em vez de `400`) quando o `username` já existe. Status descrito em `../bug-fixes-report/README.md`.
- `PUT /api/v1/order/{id}` agora **atualiza** a `VendorComission` existente em vez de inserir uma nova linha duplicada. Ver `../bug-fixes-report/001-duplicate-commission-on-order-update.md`.

Observações:
- Os nomes dos campos abaixo são exatamente os nomes usados pela API.
- Quando a rota não usa um DTO específico, ela usa o próprio modelo da entidade.

## `GET /`

Auth: livre  
Query params: nenhum  
Request body: sem corpo  
Response body: sem corpo, redireciona para `/docs/routes.md`

## `GET /health`

Auth: livre  
Query params: nenhum  
Request body: sem corpo  
Response body:

```json
{
  "status": "healthy",
  "timestamp": "2026-04-28T12:00:00Z"
}
```

## `POST /api/v1/auth/token`

Auth: livre  
Query params: nenhum  
Request body: `LoginDto`

```json
{
  "username": "string",
  "password": "string"
}
```

Response body: `TokenResponseDto`

```json
{
  "token": "string",
  "expiresAt": "2026-04-28T12:00:00Z",
  "tokenType": "Bearer"
}
```

## `GET /api/v1/car`

Auth: token (`Admin`, `Vendor`)  
Query params: nenhum  
Request body: sem corpo  
Response body: `Car[]`

```json
[
  {
    "id": 1,
    "new": true,
    "brand": "string",
    "model": "string",
    "year": 2024,
    "price": 100000.0,
    "color": "string",
    "km": 0,
    "description": "string",
    "images": [
      {
        "id": 1,
        "url": "string",
        "carId": 1
      }
    ]
  }
]
```

## `GET /api/v1/car/{id}`

Auth: token (`Admin`, `Vendor`)  
Query params: nenhum  
Request body: sem corpo  
Response body: `Car`

```json
{
  "id": 1,
  "new": true,
  "brand": "string",
  "model": "string",
  "year": 2024,
  "price": 100000.0,
  "color": "string",
  "km": 0,
  "description": "string",
  "images": [
    {
      "id": 1,
      "url": "string",
      "carId": 1
    }
  ]
}
```

## `POST /api/v1/car`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: `CreateOrUpdateCarDto`

```json
{
  "id": 1,
  "new": true,
  "brand": "string",
  "model": "string",
  "year": 2024,
  "price": 100000.0,
  "color": "string",
  "km": 0,
  "description": "string",
  "images": [
    {
      "id": 1,
      "url": "string"
    }
  ]
}
```

Response body: `Car`

```json
{
  "id": 1,
  "new": true,
  "brand": "string",
  "model": "string",
  "year": 2024,
  "price": 100000.0,
  "color": "string",
  "km": 0,
  "description": "string",
  "images": [
    {
      "id": 1,
      "url": "string",
      "carId": 1
    }
  ]
}
```

## `PUT /api/v1/car/{id}`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: `CreateOrUpdateCarDto`

```json
{
  "id": 1,
  "new": true,
  "brand": "string",
  "model": "string",
  "year": 2024,
  "price": 100000.0,
  "color": "string",
  "km": 0,
  "description": "string",
  "images": [
    {
      "id": 1,
      "url": "string"
    }
  ]
}
```

Response body: `Car`

```json
{
  "id": 1,
  "new": true,
  "brand": "string",
  "model": "string",
  "year": 2024,
  "price": 100000.0,
  "color": "string",
  "km": 0,
  "description": "string",
  "images": [
    {
      "id": 1,
      "url": "string",
      "carId": 1
    }
  ]
}
```

## `DELETE /api/v1/car/{id}`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: sem corpo  
Response body: sem corpo

## `GET /api/v1/comission`

Auth: token (`Admin`, `Vendor`)  
Query params: `vendorId?: long`  
Request body: sem corpo  
Response body: `VendorComission[]`

```json
[
  {
    "id": 1,
    "vendorId": 1,
    "vendorName": "string",
    "comissionPercentage": 3.0,
    "comissionAmount": 3000.00,
    "orderId": 1,
    "orderTotal": 100000.00
  }
]
```

## `GET /api/v1/comission/{id}`

Auth: token (`Admin`, `Vendor`)  
Query params: nenhum  
Request body: sem corpo  
Response body: `VendorComission`

```json
{
  "id": 1,
  "vendorId": 1,
  "vendorName": "string",
  "comissionPercentage": 3.0,
  "comissionAmount": 3000.00,
  "orderId": 1,
  "orderTotal": 100000.00
}
```

## `POST /api/v1/comission`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: `VendorComission`

```json
{
  "id": 1,
  "vendorId": 1,
  "vendorName": "string",
  "comissionPercentage": 3.0,
  "comissionAmount": 3000.00,
  "orderId": 1,
  "orderTotal": 100000.00
}
```

Response body: `VendorComission`

```json
{
  "id": 1,
  "vendorId": 1,
  "vendorName": "string",
  "comissionPercentage": 3.0,
  "comissionAmount": 3000.00,
  "orderId": 1,
  "orderTotal": 100000.00
}
```

## `PUT /api/v1/comission/{id}`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: `VendorComission`

```json
{
  "id": 1,
  "vendorId": 1,
  "vendorName": "string",
  "comissionPercentage": 3.0,
  "comissionAmount": 3000.00,
  "orderId": 1,
  "orderTotal": 100000.00
}
```

Response body: `VendorComission`

```json
{
  "id": 1,
  "vendorId": 1,
  "vendorName": "string",
  "comissionPercentage": 3.0,
  "comissionAmount": 3000.00,
  "orderId": 1,
  "orderTotal": 100000.00
}
```

## `DELETE /api/v1/comission/{id}`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: sem corpo  
Response body: sem corpo

## `GET /api/v1/order`

Auth: token (`Admin`, `Vendor`)  
Query params: `vendorId?: int`  
Request body: sem corpo  
Response body: `Order[]`

```json
[
  {
    "id": 1,
    "customerName": "string",
    "orderDate": "2026-04-28T12:00:00Z",
    "total": 100000.00,
    "vendorId": 1,
    "items": [
      {
        "id": 1,
        "orderId": 1,
        "carId": 1,
        "price": 100000.00,
        "discount": 5000.00
      }
    ]
  }
]
```

## `GET /api/v1/order/{id}`

Auth: token (`Admin`, `Vendor`)  
Query params: nenhum  
Request body: sem corpo  
Response body: `Order`

```json
{
  "id": 1,
  "customerName": "string",
  "orderDate": "2026-04-28T12:00:00Z",
  "total": 100000.00,
  "vendorId": 1,
  "items": [
    {
      "id": 1,
      "orderId": 1,
      "carId": 1,
      "price": 100000.00,
      "discount": 5000.00
    }
  ]
}
```

## `POST /api/v1/order`

Auth: token (`Admin`, `Vendor`)  
Query params: nenhum  
Request body: `CreateOrUpdateOrderDto`

```json
{
  "id": 1,
  "customerName": "string",
  "orderDate": "2026-04-28T12:00:00Z",
  "total": 100000.00,
  "vendorId": 1,
  "items": [
    {
      "id": 1,
      "carId": 1,
      "price": 100000.00,
      "discount": 5000.00
    }
  ]
}
```

Response body: `Order`

```json
{
  "id": 1,
  "customerName": "string",
  "orderDate": "2026-04-28T12:00:00Z",
  "total": 100000.00,
  "vendorId": 1,
  "items": [
    {
      "id": 1,
      "orderId": 1,
      "carId": 1,
      "price": 100000.00,
      "discount": 5000.00
    }
  ]
}
```

## `PUT /api/v1/order/{id}`

Auth: token (`Admin`, `Vendor`)  
Query params: nenhum  
Request body: `CreateOrUpdateOrderDto`

```json
{
  "id": 1,
  "customerName": "string",
  "orderDate": "2026-04-28T12:00:00Z",
  "total": 100000.00,
  "vendorId": 1,
  "items": [
    {
      "id": 1,
      "carId": 1,
      "price": 100000.00,
      "discount": 5000.00
    }
  ]
}
```

Response body: `Order`

```json
{
  "id": 1,
  "customerName": "string",
  "orderDate": "2026-04-28T12:00:00Z",
  "total": 100000.00,
  "vendorId": 1,
  "items": [
    {
      "id": 1,
      "orderId": 1,
      "carId": 1,
      "price": 100000.00,
      "discount": 5000.00
    }
  ]
}
```

## `DELETE /api/v1/order/{id}`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: sem corpo  
Response body: sem corpo

## `GET /api/v1/user`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: sem corpo  
Response body: `User[]`

```json
[
  {
    "id": 1,
    "username": "string",
    "password": "string",
    "email": "string",
    "comissionPerSaleInPercent": 3,
    "role": "Admin"
  }
]
```

## `GET /api/v1/user/{id}`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: sem corpo  
Response body: `User`

```json
{
  "id": 1,
  "username": "string",
  "password": "string",
  "email": "string",
  "comissionPerSaleInPercent": 3,
  "role": "Admin"
}
```

## `POST /api/v1/user`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: `User`

```json
{
  "id": 1,
  "username": "string",
  "password": "string",
  "email": "string",
  "comissionPerSaleInPercent": 3,
  "role": "Admin"
}
```

Response body: `User`

```json
{
  "id": 1,
  "username": "string",
  "password": "string",
  "email": "string",
  "comissionPerSaleInPercent": 3,
  "role": "Admin"
}
```

## `PUT /api/v1/user/{id}`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: `User`

```json
{
  "id": 1,
  "username": "string",
  "password": "string",
  "email": "string",
  "comissionPerSaleInPercent": 3,
  "role": "Admin"
}
```

Response body: `User`

```json
{
  "id": 1,
  "username": "string",
  "password": "string",
  "email": "string",
  "comissionPerSaleInPercent": 3,
  "role": "Admin"
}
```

## `DELETE /api/v1/user/{id}`

Auth: token (`Admin`)  
Query params: nenhum  
Request body: sem corpo  
Response body: sem corpo
