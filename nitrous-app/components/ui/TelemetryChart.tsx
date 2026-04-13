"use client"
import { Area, AreaChart, ResponsiveContainer, YAxis } from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

const mockData = Array.from({ length: 15 }, (_, i) => ({
  val: Math.floor(Math.random() * (8500 - 6000) + 6000),
}))

export function TelemetryChart({ title, color }: { title: string; color: string }) {
  return (
    <Card className="bg-[#0a0c10] border-zinc-800/50 shadow-2xl">
      <CardHeader className="p-3 pb-0">
        <CardTitle className="text-[10px] tracking-widest text-zinc-500 uppercase font-mono">{title}</CardTitle>
      </CardHeader>
      <CardContent className="p-0 h-24">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={mockData}>
            <defs>
              <linearGradient id={`grad-${title}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor={color} stopOpacity={0.4}/>
                <stop offset="95%" stopColor={color} stopOpacity={0}/>
              </linearGradient>
            </defs>
            <Area 
              type="stepAfter" 
              dataKey="val" 
              stroke={color} 
              fill={`url(#grad-${title})`} 
              strokeWidth={1.5}
              isAnimationActive={false} 
            />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  )
}