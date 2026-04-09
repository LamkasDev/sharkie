#version 450

layout(push_constant) uniform PC {
    float time;
} pc;
layout(location = 0) in vec2 fragPos;
layout(location = 0) out vec4 outColor;

void main() {
    float slowTime = pc.time * 0.5;
    vec3 rainbow = 0.5 + 0.5 * cos(slowTime + fragPos.xyx * 2.0 + vec3(0, 2, 4));

    outColor = vec4(rainbow, 1.0);
}