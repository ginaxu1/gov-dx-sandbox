# Portal template

This is a React + TypeScript + TailwindCSS + env File configured template using `vite.dev`.

This following guide explains how to set up and run a new portal with this template.

## Setup

1. Copy the template directory
```bash
cp -R ./template ./new-portal-name
```

2. Go into the `new-portal-name` directory
```bash
cd /new-portal-name
```

4. Install dependencies
```bash
npm install
```

4. Set up environment variables
```bash
cp .env.template .env
```
- set up the port and base_path

## Development

To run the project in development mode:

```bash
npm run dev
```

The application will be available at `http://localhost:5173`

## Code Quality

Before committing changes, run the linting check:

```bash
npm run lint
```
