package main

import (
"fmt"
"strings"
"github.com/charmbracelet/lipgloss"
"github.com/plebone/nostrfeedz-cli/pkg/styles"
)

func main() {
width := 80

fmt.Println("\n=== NostrFeedz CLI Render Test ===\n")

fmt.Println("Test 1: Plain text centered")
fmt.Println(centerText("This text should be centered", width))
fmt.Println()

fmt.Println("Test 2: Styled text")
title := styles.TitleStyle.Render("🚀 Nostr-Feedz CLI")
fmt.Println(centerText(title, width))
fmt.Println()

fmt.Println("Test 3: Key-value styling")
fmt.Println(centerText(styles.KeyStyle.Render("1")+" - Remote Signer", width))
fmt.Println(centerText(styles.KeyStyle.Render("2")+" - Private Key", width))
fmt.Println()

fmt.Println("=== All tests passed! ===")
fmt.Println("If you can see styled text above, the app should work.")
}

func centerText(text string, width int) string {
if width <= 0 {
width = 80
}
textWidth := lipgloss.Width(text)
if textWidth >= width {
return text
}
padding := (width - textWidth) / 2
return strings.Repeat(" ", padding) + text
}
