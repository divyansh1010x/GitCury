package main

// import (
// 	"GitCury/utils"
// 	"time"
// )

// func main() {
// 	// Test different animations with varying message lengths

// 	// Test 1: Spinner with short messages
// 	utils.StartCreativeLoader("Testing", utils.SpinnerAnimation)
// 	time.Sleep(2 * time.Second)
// 	utils.UpdateCreativeLoaderMessage("Short msg")
// 	time.Sleep(2 * time.Second)
// 	utils.UpdateCreativeLoaderMessage("Very long message that should be handled properly")
// 	time.Sleep(2 * time.Second)
// 	utils.ShowCompletionMessage("Spinner test completed", true)

// 	time.Sleep(1 * time.Second)

// 	// Test 2: Git animation with phase changes
// 	utils.StartCreativeLoader("Processing files", utils.GitAnimation)
// 	utils.UpdateCreativeLoaderPhase("analyzing")
// 	time.Sleep(2 * time.Second)
// 	utils.UpdateCreativeLoaderPhase("generating")
// 	time.Sleep(2 * time.Second)
// 	utils.UpdateCreativeLoaderPhase("finalizing")
// 	time.Sleep(2 * time.Second)
// 	utils.ShowCompletionMessage("Git animation test completed", true)

// 	time.Sleep(1 * time.Second)

// 	// Test 3: Braille animation
// 	utils.StartCreativeLoader("Clustering files", utils.BrailleAnimation)
// 	utils.UpdateCreativeLoaderPhase("clustering")
// 	time.Sleep(3 * time.Second)
// 	utils.UpdateCreativeLoaderMessage("Different message length")
// 	time.Sleep(2 * time.Second)
// 	utils.ShowCompletionMessage("Braille animation test completed", true)

// 	// Test 4: Dots animation
// 	utils.StartCreativeLoader("Loading", utils.DotsAnimation)
// 	time.Sleep(3 * time.Second)
// 	utils.UpdateCreativeLoaderMessage("This is a much longer message to test padding")
// 	time.Sleep(2 * time.Second)
// 	utils.ShowCompletionMessage("Dots animation test completed", true)

// 	// Test error case
// 	time.Sleep(1 * time.Second)
// 	utils.StartCreativeLoader("Testing error", utils.ProcessingAnimation)
// 	time.Sleep(2 * time.Second)
// 	utils.ShowCompletionMessage("Error test - something went wrong", false)
// }
