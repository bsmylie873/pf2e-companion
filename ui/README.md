# PF2E Companion — Frontend

The web frontend for the PF2E Companion application, built with React, TypeScript, and Vite.

## Tech Stack

| Technology | Purpose |
|---|---|
| [React](https://react.dev/) 19 | UI framework |
| [TypeScript](https://www.typescriptlang.org/) 6 | Type safety |
| [Vite](https://vite.dev/) 8 | Build tool and dev server |
| [React Router](https://reactrouter.com/) 7 | Client-side routing |
| [TipTap](https://tiptap.dev/) 3 | Rich-text editor (session notes, personal notes) |
| [Motion](https://motion.dev/) | Animations |
| [Vitest](https://vitest.dev/) | Unit testing |
| [Testing Library](https://testing-library.com/) | Component testing utilities |

## Prerequisites

- [Node.js](https://nodejs.org/) >= 20
- npm (bundled with Node.js)
- A running backend API (see the [root README](../README.md) for setup)

## Setup

```bash
# Install dependencies
npm install

# Start the development server (http://localhost:5173)
npm run dev
```

The dev server expects the backend API at `http://localhost:8080` by default. Override this with the `VITE_API_BASE_URL` environment variable.

## Available Scripts

| Script | Description |
|---|---|
| `npm run dev` | Start the Vite dev server with HMR |
| `npm run build` | Type-check and build for production |
| `npm run preview` | Preview the production build locally |
| `npm run lint` | Run ESLint |
| `npm test` | Run tests once |
| `npm run test:watch` | Run tests in watch mode |
| `npm run coverage` | Run tests with coverage report |

## Key Features

- **Campaign dashboard** — view and manage all campaigns on the games list
- **Game editor** — tabbed interface for sessions, notes, characters, items, and folders
- **Session notes** — rich-text collaborative editor with tables, task lists, highlights, images, typography, and alignment
- **Interactive map view** — upload campaign maps, place/drag session pins, manage pin groups
- **Personal & shared notes** — folder-organised note system with the same rich-text editor
- **Character sheets** — PF2E stat blocks for PCs and NPCs
- **Item management** — inventory tracking with traits, bulk, and pricing
- **Real-time updates** — WebSocket connection for live collaboration and event broadcasting
- **Magic link invites** — join campaigns via shareable invite links
- **Patch notes** — in-app release notes rendered from `RELEASE_NOTES.md`
- **Password reset** — forgot/reset password flow

## Project Structure

```
ui/
├── src/
│   ├── api/          # API client functions
│   ├── components/   # Reusable UI components
│   ├── constants/    # Application constants
│   ├── context/      # React context providers (Auth, MapNav)
│   ├── hooks/        # Custom hooks (e.g. useGameSocket)
│   ├── pages/        # Route-level page components
│   ├── types/        # TypeScript type definitions
│   ├── App.tsx       # Root component with route definitions
│   └── main.tsx      # Entry point
├── public/           # Static assets
├── index.html        # HTML entry point
├── vite.config.ts    # Vite configuration
├── tsconfig.json     # TypeScript config
└── package.json
```

## Contributing

1. Ensure the backend and database are running (see the [root README](../README.md))
2. Create a feature branch from `master`
3. Run `npm run lint` and `npm test` before committing
4. All tests must pass with ≥ 80% statement coverage (`npm run coverage`)
5. Open a pull request targeting `master`

## License

This project is licensed under the MIT License — see the [LICENSE](../LICENSE) file for details.
