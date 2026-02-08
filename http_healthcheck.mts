// Only used with bun -e, not built for script usage
const r = await fetch("http://localhost:3000/chartmaker/ping");
if (r.status !== 200) {
    throw new Error("status not 200")
}
console.log(r.status)