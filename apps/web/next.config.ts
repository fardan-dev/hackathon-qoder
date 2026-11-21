import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  poweredByHeader: false,
  output: "standalone",
  generateBuildId: async () => process.env.GIT_SHA ?? "local",
};

export default nextConfig;
