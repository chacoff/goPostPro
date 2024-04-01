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

function extractData(start_timestamp::TimeDate, end_timestamp::TimeDate)

    dff = filter(row -> row.Times >= start_timestamp && row.Times <= end_timestamp, df)
    # dff_x = dff.Temps    
    # heatmap(hcat(dff_x...))

    return dff
end

filename = "D:\\00_Dev\\GoPostPro\\juliaPostPro\\DUO01-02_0891_half.txt"
stamps, temps = readFile(filename)

df = DataFrame(Times=stamps, Temps=temps)
println(df)
dff = extractData(TimeDate("2024-02-13T11:05:06.100"), TimeDate("2024-02-13T11:05:16.000"))
println(dff)
println("Ok!")


