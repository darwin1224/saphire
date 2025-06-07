# Saphire

A minimalistic Turing-complete functional programming language designed with tree-walk interpreter.

# Installation

Make sure to have `go` installed:

```bash
go install github.com/darwin1224/saphire@latest
```

Then run

```bash
saphire your_code.sp

# or for REPL

saphire
```

# Features

- First-class functions
- Closures
- Conditional Flow
- Recursion
- Dynamic Typing
- Strong Typing
- Automatic memory management
- String interpolation
- Array built-ins
- Hash map built-ins

# TODO Features

- Module System
- Namespaces
- Standard Library
- Reflection
- Native Concurrency
- VM-based or Bytecode Interpreter
- JIT Interpreter

# Examples

## Generate $\pi$

To approximate $\pi$, we use Nilakantha series which given by following:

$$
\pi = 3 + \sum_{n=1}^{\infty} (-1)^{n+1} \frac{4}{(2n)(2n+1)(2n+2)}.
$$

This expression starts with 3 and alternates between adding and subtracting fractions involving the product of three consecutive even integers. Written out, the series appears as:

$$
\pi = 3 + \frac{4}{2 \cdot 3 \cdot 4} - \frac{4}{4 \cdot 5 \cdot 6} + \frac{4}{6 \cdot 7 \cdot 8} - \cdots
$$

For sufficiently large $N \in \mathbb{N}$, we approximate the series to a finite number of terms:

$$
\begin{aligned}
\pi &\approx 3 + \sum_{n=1}^{N} (-1)^{n+1} \frac{4}{(2n)(2n+1)(2n+2)} \\
&= 3 + \frac{4}{2 \cdot 3 \cdot 4} - \frac{4}{4 \cdot 5 \cdot 6} + \frac{4}{6 \cdot 7 \cdot 8} - \cdots \pm \frac{4}{(2N)(2N+1)(2N+2)} \\
&\approx 3.141 \dots
\end{aligned}
$$

Thus, for sufficiently large $N$, the partial sum

$$
3 + \sum_{n=1}^{N} (-1)^{n+1} \frac{4}{(2n)(2n+1)(2n+2)}
$$

should give a close numerical approximation to $\pi$.

Example code with $N = 1000$:

```
// Nilakantha series approximation of π with N = 1000.

let nilakanthaTerm = fn(n) {
  (-1) ** (n+1) * 4 / ((2*n) * (2*n+1) * (2*n+2))
}

let nilakanthaPi = fn(n, i) {
  if (i <= n) {
    nilakanthaTerm(i) + nilakanthaPi(n, i+1)
  } else {
    0
  }
}

let generatePi = fn(n) {
  3 + nilakanthaPi(n, 1)
}

let N = 1000;
let pi = generatePi(N)

print(pi)

// Output: 3.14159265334054182972
```

## Generate Euler number $e$

Generally, the Taylor series of $f(x) = e^x$ where $f: \mathbb{R} \to \mathbb{R}$ is given by:

```math
e^x = \sum_{n=0}^{\infty} \frac{x^n}{n!}, \,\,\,\,\, \text{for all } x \in \mathbb{R}.
```

To approximate Euler number $e$, we evaluate this at $x = 1$. Then we have $e^1 = e$ with:

```math
e = \sum_{n=0}^{\infty} \frac{1^n}{n!} = \sum_{n=0}^{\infty} \frac{1}{n!}.
```

For sufficiently large $N \in \mathbb{N}$, we approximate the series to some finite $N$:

```math
\begin{align*}
e &\approx \sum_{n=0}^{N} \frac{1}{n!} \\
&= 1 + 1 + \frac{1}{2} + \frac{1}{6} + \frac{1}{24} + \frac{1}{120} + \cdots \\
&\approx 2.718 \dots
\end{align*}
```

Thereby, for sufficiently large $N$, the partial sum $\sum_{n=0}^N \frac{1}{n!}$ should give a close numerical approximation to $e$.

Example code with $N = 1000$:

```
// Taylor series approximation of Euler number with N = 1000.

let factorial = fn(n) {
  if (n == 0) {
    1
  } else {
    n * factorial(n-1)
  }
}

let taylorE = fn(n, i) {
  if (i <= n) {
    (1 / factorial(i)) + taylorE(n, i+1)
  } else {
    0
  }
}

let generateE = fn(n) {
  taylorE(n, 0)
}

let N = 1000;
let e = generateE(N)

print(e)

// Output: 2.71828182845904509080
```