<template>
    <div
        class="comment-content text-foreground mb-4 leading-relaxed pl-[52px] text-sm break-words transition-all duration-300">
        <slot />
    </div>
</template>

<style scoped>
.comment-content :deep(img) {
    max-width: 64px;
    height: auto;
    border-radius: 0.3rem;
    margin: 0.5rem 0.5rem;
    display: inline-block;
    cursor: zoom-in;
    transition: transform 0.2s;
}

.comment-content :deep(img):hover {
    transform: scale(1.05);
}

/* Specific fix for standard emoji size when rendered as img */
.comment-content :deep(.wl-emoji) {
    display: inline-block;
    vertical-align: middle;
    height: 1.5em !important;
    width: auto !important;
    margin: 0 0.1em;
    border-radius: 0;
    cursor: default;
}

/* Markdown Elements Styling */
.comment-content :deep(p) {
    margin-bottom: 0.5em;
}

.comment-content :deep(p:last-child) {
    margin-bottom: 0;
}

.comment-content :deep(a) {
    color: hsl(var(--primary));
    text-decoration: underline;
    text-underline-offset: 4px;
}

.comment-content :deep(blockquote) {
    display: block;
    border-left: 4px solid hsl(var(--primary));
    background-color: hsl(var(--muted)/0.3);
    padding: 0.5em 1em;
    margin: 1em 0;
    color: hsl(var(--muted-foreground));
    border-radius: 0 0.3rem 0.3rem 0;
    quotes: none;
}

.comment-content :deep(pre) {
    background-color: hsl(var(--secondary)/0.5);
    padding: 0.5em;
    border-radius: 0.5rem;
    overflow-x: auto;
    margin: 0.5em 0;
}

.comment-content :deep(code) {
    background-color: hsl(var(--secondary)/0.5);
    padding: 0.1em 0.3em;
    border-radius: 0.2rem;
    font-family: monospace;
    font-size: 0.9em;
}

.comment-content :deep(pre) :deep(code) {
    background-color: transparent;
    padding: 0;
}

/* Headings */
.comment-content :deep(h1),
.comment-content :deep(h2),
.comment-content :deep(h3),
.comment-content :deep(h4),
.comment-content :deep(h5),
.comment-content :deep(h6) {
    font-weight: bold;
    line-height: 1.25;
    margin-top: 1em;
    margin-bottom: 0.5em;
    color: hsl(var(--foreground));
}

.comment-content :deep(h1) {
    font-size: 1.4em;
    padding-bottom: 0.3em;
    border-bottom: 1px solid hsl(var(--border));
}

.comment-content :deep(h2) {
    font-size: 1.25em;
    padding-bottom: 0.3em;
    border-bottom: 1px solid hsl(var(--border));
}

.comment-content :deep(h3) {
    font-size: 1.1em;
}

.comment-content :deep(h4) {
    font-size: 1em;
}

.comment-content :deep(h5) {
    font-size: 0.9em;
}

.comment-content :deep(h6) {
    font-size: 0.85em;
    color: hsl(var(--muted-foreground));
}

/* Lists */
.comment-content :deep(ul),
.comment-content :deep(ol) {
    margin-left: 1.5em;
    margin-bottom: 1em;
}

.comment-content :deep(ul) {
    list-style-type: disc;
}

.comment-content :deep(ol) {
    list-style-type: decimal;
}

.comment-content :deep(li) {
    margin-top: 0.25em;
}

/* Nested Lists */
.comment-content :deep(ul ul),
.comment-content :deep(ul ol),
.comment-content :deep(ol ul),
.comment-content :deep(ol ol) {
    margin-bottom: 0;
    margin-top: 0;
}

/* Horizontal Rule */
.comment-content :deep(hr) {
    height: 0.25em;
    padding: 0;
    margin: 1.5em 0;
    background-color: hsl(var(--border));
    border: 0;
}

/* Tables */
.comment-content :deep(table) {
    border-spacing: 0;
    border-collapse: collapse;
    margin-top: 0;
    margin-bottom: 1em;
    width: 100%;
    overflow: auto;
    display: block;
}

.comment-content :deep(table tr) {
    background-color: transparent;
    border-top: 1px solid hsl(var(--muted));
}

.comment-content :deep(table tr:nth-child(2n)) {
    background-color: hsl(var(--muted)/0.2);
}

.comment-content :deep(table th),
.comment-content :deep(table td) {
    padding: 0.5em 1em;
    border: 1px solid hsl(var(--border));
}

.comment-content :deep(table th) {
    font-weight: 600;
    background-color: hsl(var(--muted)/0.5);
    text-align: left;
}

/* Task Lists */
.comment-content :deep(.contains-task-list) {
    list-style-type: none;
    padding-left: 0;
    margin-left: 0;
}

.comment-content :deep(.task-list-item) {
    display: flex;
    align-items: center;
    gap: 0.5em;
}

.comment-content :deep(input[type="checkbox"]) {
    margin: 0;
    accent-color: hsl(var(--primary));
}

/* Links Hover Effect */
.comment-content :deep(a:hover) {
    color: hsl(var(--primary)/0.8);
    text-decoration-thickness: 2px;
}

/* Sticker handling */
.comment-content :deep(img[src*="sticker"]),
.comment-content :deep(img[class*="sticker"]),
.comment-content :deep(img[src*="emojis"]),
.comment-content :deep(img[src*="bilibili"]),
.comment-content :deep(img[src*="sinaimg"]),
.comment-content :deep(img[src*="qzone"]),
.comment-content :deep(img[src*="baidu"]),
.comment-content :deep(img[src*="akamaized"]),
.comment-content :deep(img[alt*="sticker"]) {
    border-radius: 0;
    margin: 0;
    max-width: 150px;
    max-height: 150px;
}
</style>
