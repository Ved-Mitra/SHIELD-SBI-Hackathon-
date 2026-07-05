// ContentView.swift — SHIELD iOS ContentView stub
// Calls into the shared KMP module via Swift/Kotlin interop.

import SwiftUI
// import shared  ← uncomment once XcodeGen/KMP Xcode integration is configured

struct ContentView: View {
    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "shield.fill")
                .resizable()
                .frame(width: 80, height: 80)
                .foregroundColor(.blue)
            Text("SHIELD Guardian")
                .font(.largeTitle)
                .bold()
            Text("On-Device Threat Scanner")
                .font(.subheadline)
                .foregroundColor(.secondary)
            Text("iOS support coming soon.\nURL Risk Engine & APK Scanner\nare Android-native features.")
                .multilineTextAlignment(.center)
                .foregroundColor(.secondary)
                .padding()
        }
        .padding()
    }
}

#Preview {
    ContentView()
}
