# using Pkg
# Pkg.add("DataFrames")
# Pkg.add("TimeZones")
# Pkg.add("CompoundPeriods")
# Pkg.add("TimesDates")
# precompile
using TimesDates, Dates
using DataFrames
using Plots

function readFile(filename::AbstractString)
    
    timestamps = []
    temperatures = []
    
    file = open(filename, "r")
    readline(file)
    
    try
        while !eof(file)
            line = readline(file)  # skip the first line
            line = replace(line, "," => ".")
            elements = split(line)
            
            if length(elements) >= 2    
                times = join(elements[1:2], "T")
                timeStamp = TimeDate(times)

                temps = join(elements[8:end-4], " ")  # end-4
                tempsArray = [parse(Float64, substr) for substr in split(temps)]

                push!(timestamps, timeStamp)
                push!(temperatures, tempsArray)
            end
        end
    finally
        close(file)
    end

    return timestamps, temperatures
end


filename = "DUO01-02_0891_half.txt"
stamps, temps = readFile(filename)

df = DataFrame(Times=stamps, Temps=temps)
println(df)

start_timestamp = TimeDate("2024-02-13T11:05:06.600")
end_timestamp = TimeDate("2024-02-13T11:05:06.957")

dff = filter(row -> row.Times >= start_timestamp && row.Times <= end_timestamp, df)
dff_x = dff.Temps
# dff_cat = hcat(dff_x...)
dff_y = repeat(1:length(dff.Temps), 1)
dff_z = [400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500]

heatmap(hcat(dff_x...))