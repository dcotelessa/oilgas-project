# Frontend - Oil & Gas Inventory System

Vue.js 3 frontend with TypeScript, Vite, and Pinia for state management.

## Structure

```
frontend/
├── src/
│   ├── components/    # Reusable components
│   ├── views/        # Page components
│   ├── stores/       # Pinia stores
│   └── utils/        # Utility functions
├── public/           # Static assets
└── index.html        # Entry point
```

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Type checking
npm run type-check
```

## Features

- **Dashboard**: Overview of inventory and operations
- **Customers**: Customer management
- **Inventory**: Product inventory tracking
- **Grades**: Oil & gas grade management (J55, JZ55, L80, N80, P105, P110)

## State Management

Uses Pinia for state management with TypeScript support.

## Styling

Component-scoped CSS with modern responsive design.
